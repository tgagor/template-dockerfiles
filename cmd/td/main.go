package main

import (
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/tgagor/template-dockerfiles/pkg/config"
	"github.com/tgagor/template-dockerfiles/pkg/parser"
	"github.com/tgagor/template-dockerfiles/pkg/util"
)

var BuildVersion string // Will be set dynamically at build time.
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
			fmt.Printf("%s version: %s\n", appName, BuildVersion)
			return
		}

		// Main logic goes here
		if flags.Verbose {
			slog.Debug("Verbose mode enabled.")
		}
		// if flags.DryRun {
		// 	slog.Info("Dry run enabled - no actions will be executed.")
		// }
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
			util.FailOnError(err, "Error during parsing")
		}
	},
}

func init() {
	if BuildVersion == "" {
		BuildVersion = "development" // Fallback if not set during build
	}

	cmd.PersistentFlags().StringVarP(&flags.BuildFile, "config", "c", "", "Path to the configuration file (required)")
	// rootCmd.MarkPersistentFlagRequired("config")

	cmd.Flags().BoolVarP(&flags.Build, "build", "b", false, "Build Docker images after templating")
	cmd.Flags().BoolVarP(&flags.Push, "push", "p", false, "Push Docker images after building")
	cmd.Flags().BoolVarP(&flags.Delete, "delete", "d", false, "Delete templated Dockerfiles after successful building")
	// cmd.Flags().BoolVarP(&flags.DryRun, "dry-run", "d", false, "Print actions but don't execute them")
	cmd.Flags().IntVar(&flags.Threads, "parallel", runtime.NumCPU(), "Specify the number of threads to use, defaults to number of CPUs")
	cmd.Flags().StringVarP(&flags.Tag, "tag", "t", "", "Tag to use as the image version")
	cmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Increase verbosity of output")
	cmd.Flags().BoolVarP(&flags.PrintVersion, "version", "V", false, "Display the application version and exit")
}

func main() {
	if err := cmd.Execute(); err != nil {
		util.FailOnError(err)
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
