package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"td/pkg/config"
	"td/pkg/parser"
)

var Version string // Will be set dynamically at build time.
var appName string = "td"
var flags config.Flags

var cmd = &cobra.Command{
	Use:   appName,
	Short: "A Docker image builder that uses Go templates to dynamically generate Dockerfiles.",
	Long: `A CLI tool for building Docker images with configurable Dockerfile templates and multi-threaded execution.

When 'docker build' is just not enough. :-)`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config file requirement if --version is provided
		if flags.PrintVersion {
			return nil
		}
		// Validate config file if it's not provided and --version isn't invoked
		if flags.BuildFile == "" {
			return fmt.Errorf("the --config flag is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		initLogger(flags.Verbose)

		// If version flag is provided, show the version and exit.
		if flags.PrintVersion {
			fmt.Printf("%s version: %s\n", appName, Version)
			return
		}

		// Main logic goes here
		if flags.Verbose {
			slog.Debug("Verbose mode enabled.")
		}
		if flags.DryRun {
			slog.Info("Dry run enabled - no actions will be executed.")
		}
		if flags.Push {
			slog.Warn("Images will be pushed after building.")
		}
		slog.Info("Number of", "threads", flags.Threads)
		if flags.Tag != "" {
			slog.Info("Setting", "tag", flags.Tag)
		}

		// Parse configuration file
		slog.Info("Loading", "config", flags.BuildFile)
		cfg := config.Load(flags.BuildFile)
		slog.Debug("Loaded", "config", cfg)

		// Run templating and image building
		workdir := filepath.Dir(flags.BuildFile)
		if err := parser.Run(workdir, cfg, flags); err != nil {
			slog.Error("Error during parsing", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	if Version == "" {
		Version = "development" // Fallback if not set during build
	}

	cmd.PersistentFlags().StringVarP(&flags.BuildFile, "config", "c", "", "Path to the configuration file (required)")
	// rootCmd.MarkPersistentFlagRequired("config")

	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "Print actions but don't execute them")
	cmd.Flags().BoolVar(&flags.Push, "push", false, "Push Docker images after building")
	cmd.Flags().IntVarP(&flags.Threads, "parallel", "p", runtime.NumCPU(), "Specify the number of threads to use (default: number of CPUs)")
	cmd.Flags().StringVarP(&flags.Tag, "tag", "t", "", "Tag to use as the image version")
	cmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Increase verbosity of output")
	cmd.Flags().BoolVarP(&flags.PrintVersion, "version", "V", false, "Display the application version and exit")
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initLogger(verbose bool) {
	// Disable timestamp in logger
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	// Configure log level
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}
}

// func main() {
//     // Initialize flags
//     args := flags.Parse()

// 	fmt.Println(args)
// 	if Version == "" {
//         Version = "development" // Fallback if not set during build
//     }
//     fmt.Println("Version:", Version)

//     // // Configure logger
//     // log := logger.New(args.Verbose)

//     // // Parse configuration file
//     // cfg, err := config.Load(args.ConfigFile)
//     // if err != nil {
//     //     log.Fatal("Error loading config:", err)
//     // }

//     // // Run templating and image building
//     // if err := parser.Run(cfg, log); err != nil {
//     //     log.Fatal("Error during parsing:", err)
//     // }

//     // // Execute tasks in parallel
//     // runner.ExecuteTasks(cfg, log)
// }
