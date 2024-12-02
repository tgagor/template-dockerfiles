package main

import (
	"fmt"
	"os"
	"github.com/jessevdk/go-flags"
	"log"
)

var (
	// Define a struct that represents the command-line flags
	options struct {
		ConfigFile string `short:"c" long:"config" description:"Path to the configuration file" required:"true"`
		DryRun     bool   `long:"dry-run" description:"Print what would be done, but don't do anything"`
		Push       bool   `long:"push" description:"Push Docker images when successfully built"`
		Threads    int    `long:"parallel" description:"Specify the number of threads to use (default: number of CPUs)" default:"1"`
		LogLevel   string `short:"v" long:"verbose" description:"Set log level" choice:"DEBUG" choice:"INFO" default:"INFO"`
		Version    bool   `long:"version" description:"Show the version of the application and exit"`
		Tag        string `short:"t" long:"tag" description:"Tag that could be used as an image version" required:"true"`
	}
)

func main() {
	// Initialize the flag parser
	parser := flags.NewParser(&options, flags.Default)

	// Parse the flags
	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			// Show help if there's an error
			return
		}
		// If error occurs, exit with status 1
		fmt.Println("Error parsing flags:", err)
		os.Exit(1)
	}

	// Simulating your flag handling logic (you can replace with actual logic)
	if options.DryRun {
		fmt.Println("Dry run enabled")
	}

	// Print the values of flags to show they are parsed correctly
	fmt.Println("Config File:", options.ConfigFile)
	fmt.Println("Push:", options.Push)
	fmt.Println("Threads:", options.Threads)
	fmt.Println("Log Level:", options.LogLevel)
	fmt.Println("Tag:", options.Tag)

	// You can add any specific validation functions as needed, such as validating threads, etc.
	// For example, you could check if threads is a valid number
	if options.Threads <= 0 {
		log.Fatal("The number of threads must be greater than 0")
	}

}
