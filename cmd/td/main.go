package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
			// v0.6.3 (go1.23.4 on darwin/arm64; gc)
			fmt.Printf("%s (%s on %s/%s; %s)\n", BuildVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH, runtime.Compiler)
			return
		}

		// Main logic goes here
		if flags.Verbose {
			log.Debug().Msg("Verbose mode enabled.")
		}
		if flags.Push && !flags.Build {
			log.Warn().Msg("Attempting to push images without building. Ensure images were built previously, otherwise the push will fail.")
		}
		if flags.Image != "" {
			log.Warn().Str("image", flags.Image).Msg("Limiting build to a single image")
		}
		if flags.Build {
			log.Info().Msg("Images will be build after templating.")
		}
		if flags.Push {
			log.Warn().Msg("Images will be pushed after building.")
		}
		if flags.Delete {
			log.Warn().Msg("Templated Dockerfiles will be deleted at end.")
		}
		log.Info().Int("threads", flags.Threads).Msg("Number of")
		if flags.Tag != "" {
			log.Info().Str("tag", flags.Tag).Msg("Setting")
		}

		// Parse configuration file
		log.Info().Str("config", flags.BuildFile).Msg("Loading")
		cfg, err := config.Load(flags.BuildFile)
		util.FailOnError(err)
		log.Trace().Str("config", fmt.Sprintf("%#v", cfg)).Msg("Loaded")

		// Check if the image flag is valid
		if flags.Image != "" {
			if _, ok := cfg.Images[flags.Image]; !ok {
				log.Error().Str("image", flags.Image).Msg("Image not found in configuration")
				log.Error().Interface("available", cfg.ImageOrder).Msg("Try one of the following:")
				os.Exit(1)
			}
		}

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
	cmd.Flags().StringVarP(&flags.Image, "image", "i", "", "Limit the build to a single image")
	// FIXME: add more engines, add check for correctness, podman, buildx, kaniko)")
	cmd.Flags().StringVarP(&flags.Engine, "engine", "e", "docker", "Select the container engine to use (docker, buiildx)")
	cmd.Flags().BoolVarP(&flags.Push, "push", "p", false, "Push Docker images after building")
	cmd.Flags().BoolVarP(&flags.Delete, "delete", "d", false, "Delete templated Dockerfiles after successful building")
	cmd.Flags().BoolVarP(&flags.Squash, "squash", "s", false, "Squash images to reduce size (experimental)")
	// cmd.Flags().BoolVarP(&flags.DryRun, "dry-run", "d", false, "Print actions but don't execute them")
	cmd.Flags().IntVar(&flags.Threads, "parallel", runtime.NumCPU(), "Specify the number of threads to use, defaults to number of CPUs")
	cmd.Flags().StringVarP(&flags.Tag, "tag", "t", "", "Tag to use as the image version")
	cmd.Flags().BoolVar(&flags.NoColor, "no-color", false, "Disable color output")
	cmd.Flags().BoolVarP(&flags.Verbose, "verbose", "v", false, "Increase verbosity of output")
	cmd.Flags().BoolVarP(&flags.PrintVersion, "version", "V", false, "Display the application version and exit")
}

func main() {
	if err := cmd.Execute(); err != nil {
		util.FailOnError(err)
	}
}

func initLogger(verbose bool) {
	// Console writer
	consoleWriter := zerolog.ConsoleWriter{
		Out:     colorable.NewColorableStdout(),
		NoColor: flags.NoColor,
	}
	// Disable timestamps
	zerolog.TimeFieldFormat = ""
	consoleWriter.FormatTimestamp = func(i interface{}) string {
		return ""
	}

	// Custom format for level
	// consoleWriter.FormatLevel = func(i interface{}) string {
	//     if ll, ok := i.(string); ok {
	//         switch ll {
	//         case "debug":
	//             return "\033[01;36mDEBUG\033[0m" // Cyan
	//         case "info":
	//             return "\033[32mINFO\033[0m" // Green
	//         case "warn":
	//             return "\033[33mWARN\033[0m" // Yellow
	//         case "error":
	//             return "\033[31mERROR\033[0m" // Red
	//         case "fatal":
	//             return "\033[35mFATAL\033[0m" // Magenta
	//         case "panic":
	//             return "\033[31mPANIC\033[0m" // Red
	//         default:
	//             return ll
	//         }
	//     }
	//     return ""
	// }

	// Base logger
	baseLogger := zerolog.New(consoleWriter).With().Logger()

	// Add caller only for debug level using a hook
	if verbose {
		log.Logger = baseLogger.Hook(zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
			if level == zerolog.DebugLevel {
				e.Caller()
			}
		})).Level(zerolog.DebugLevel)
	} else {
		log.Logger = baseLogger.Level(zerolog.InfoLevel)
	}
}
