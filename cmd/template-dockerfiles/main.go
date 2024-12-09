package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"template-dockerfiles/pkg/config"
	"template-dockerfiles/pkg/parser"
	// "template-dockerfiles/pkg/runner"
)

var version string // Will be set dynamically at build time.
var appName string = "td"

var (
	buildFile    string
	dryRun       bool
	push         bool
	threads      int
	tag          string
	verbose      bool
	printVersion bool
)

var cmd = &cobra.Command{
	Use:   appName,
	Short: "A Docker image builder that uses Go templates to dynamically generate Dockerfiles.",
	Long: `A CLI tool for building Docker images with configurable Dockerfile templates and multi-threaded execution.

When 'docker build' is just not enough. :-)`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config file requirement if --version is provided
		if printVersion {
			return nil
		}
		// Validate config file if it's not provided and --version isn't invoked
		if buildFile == "" {
			return fmt.Errorf("the --config flag is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		initLogger(verbose)

		// If version flag is provided, show the version and exit.
		if printVersion {
			fmt.Printf("%s version: %s\n", appName, version)
			return
		}

		// Main logic goes here
		if verbose {
			slog.Debug("Verbose mode enabled.")
		}
		if dryRun {
			slog.Info("Dry run enabled - no actions will be executed.")
		}
		if push {
			slog.Warn("Images will be pushed after building.")
		}
		slog.Info("Number of", "threads", threads)
		if tag != "" {
			slog.Info("Setting", "tag", tag)
		}

		// Parse configuration file
		slog.Info("Loading", "config", buildFile)
		cfg, err := config.Load(buildFile)
		if err != nil {
			slog.Error("Error loading config", "error", err)
		}
		slog.Debug("Loaded", "config", cfg)

		// Run templating and image building
		workdir := filepath.Dir(buildFile)
		if err := parser.Run(workdir, cfg, map[string]any{
			"tag": tag,
			"dryRun": dryRun,
			"threads": threads,
		}); err != nil {
			slog.Error("Error during parsing", "error", err)
			os.Exit(1)
		}
	},
}

func init() {
	if version == "" {
		version = "development" // Fallback if not set during build
	}

	cmd.PersistentFlags().StringVarP(&buildFile, "config", "c", "", "Path to the configuration file (required)")
	// rootCmd.MarkPersistentFlagRequired("config")

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions but don't execute them")
	cmd.Flags().BoolVar(&push, "push", false, "Push Docker images after building")
	cmd.Flags().IntVarP(&threads, "parallel", "p", runtime.NumCPU(), "Specify the number of threads to use (default: number of CPUs)")
	cmd.Flags().StringVarP(&tag, "tag", "t", "", "Tag to use as the image version")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Increase verbosity of output")
	cmd.Flags().BoolVarP(&printVersion, "version", "V", false, "Display the application version and exit")
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
