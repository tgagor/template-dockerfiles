package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	// "template-dockerfiles/pkg/config"
	// "template-dockerfiles/pkg/flags"
	// "template-dockerfiles/pkg/logger"
	// "template-dockerfiles/pkg/parser"
	// "template-dockerfiles/pkg/runner"
)

var Version string // Will be set dynamically at build time.
var (
	configFile string
	dryRun     bool
	push       bool
	parallel   int
	tag        string
	verbose    bool
	version	   bool
)

var rootCmd = &cobra.Command{
	Use:   "template-dockerfiles",
	Short: "A Docker image builder that uses Go templates to dynamically generate Dockerfiles.",
	Long:  `A CLI tool for building Docker images with configurable Dockerfile templates and multi-threaded execution.

When 'docker build' is just not enough. :-)`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config file requirement if --version is provided
		if version {
			return nil
		}
		// Validate config file if it's not provided and --version isn't invoked
		if configFile == "" {
			return fmt.Errorf("the --config flag is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// If version flag is provided, show the version and exit.
		if version {
			fmt.Printf("template-dockerfiles version: %s\n", Version)
			os.Exit(0)
		}

		// Main logic goes here
		if verbose {
			fmt.Println("Verbose mode enabled.")
		}
		fmt.Printf("Using config file: %s\n", configFile)
		if dryRun {
			fmt.Println("Dry run enabled - no actions will be executed.")
		}
		if push {
			fmt.Println("Images will be pushed after building.")
		}
		fmt.Printf("Number of threads: %d\n", parallel)
		if tag != "" {
			fmt.Printf("Using tag: %s\n", tag)
		}
	},
}

func init() {
	if Version == "" {
        Version = "development" // Fallback if not set during build
    }
	// fmt.Println("Version:", Version)

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Path to the configuration file (required)")
	rootCmd.MarkPersistentFlagRequired("config")

	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print actions but don't execute them")
	rootCmd.Flags().BoolVar(&push, "push", false, "Push Docker images after building")
	rootCmd.Flags().IntVarP(&parallel, "parallel", "p", runtime.NumCPU(), "Specify the number of threads to use (default: number of CPUs)")
	rootCmd.Flags().StringVarP(&tag, "tag", "t", "", "Tag to use as the image version")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Increase verbosity of output")
	rootCmd.Flags().BoolVarP(&version, "version", "V", false, "Display the application version and exit")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
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
