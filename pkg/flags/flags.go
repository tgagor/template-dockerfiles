package flags

import "flag"
import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

type Args struct {
    ConfigFile 	string
    Verbose    	bool
	DryRun		bool
	Push		bool
	Threads		int
	//LogLevel	string
	Version		bool
	Tag			string

	// ConfigFile string `short:"c" long:"config" description:"Path to the configuration file" required:"true"`
	// DryRun     bool   `long:"dry-run" description:"Print what would be done, but don't do anything"`
	// Push       bool   `long:"push" description:"Push Docker images when successfully built"`
	// Threads    int    `long:"parallel" description:"Specify the number of threads to use (default: number of CPUs)" default:"1"`
	// LogLevel   string `short:"v" long:"verbose" description:"Set log level" choice:"DEBUG" choice:"INFO" default:"INFO"`
	// Version    bool   `long:"version" description:"Show the version of the application and exit"`
	// Tag        string `short:"t" long:"tag" description:"Tag that could be used as an image version" required:"true"`
}

var rootCmd = &cobra.Command{
	Use:   "template-dockerfiles",
	Short: "A Docker image builder that uses Go templates to dynamically generate Dockerfiles.",
	Long:  `A CLI tool for building Docker images with configurable Dockerfile templates and multi-threaded execution.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Main logic goes here
		if verbose {
			fmt.Println("Verbose mode enabled")
		}
		fmt.Printf("Using config file: %s\n", configFile)
	},
}


func Parse() *Args {
    var args Args
    // flag.StringVar(&args.ConfigFile, "config", "build.yaml", "Path to configuration file")
    // flag.BoolVar(&args.Verbose, "verbose", false, "Enable verbose logging")
    // flag.Parse()

	rootCmd.PersistentFlags().StringVarP(&args.ConfigFile, "config", "c", "", "Path to the configuration file (required)")
	rootCmd.MarkPersistentFlagRequired("config")

	rootCmd.Flags().BoolVar(&args.DryRun, "dry-run", false, "Print actions but don't execute them")
	rootCmd.Flags().BoolVar(&args.Push, "push", false, "Push Docker images after building")
	rootCmd.Flags().IntVarP(&args.Threads, "parallel", "p", runtime.NumCPU(), "Specify the number of threads to use (default: number of CPUs)")
	rootCmd.Flags().StringVarP(&args.Tag, "tag", "t", "", "Tag to use as the image version")
	rootCmd.Flags().BoolVarP(&args.Verbose, "verbose", "v", false, "Increase verbosity of output")
	rootCmd.Flags().BoolP("version", "V", false, "Display the application version and exit")

    return &args
}






func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
