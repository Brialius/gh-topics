package cli

import (
	"fmt"
	"github.com/Brialius/gh-topics/pkg/config"
	"github.com/Brialius/gh-topics/pkg/errors"
	"github.com/Brialius/gh-topics/pkg/topics"
	flag "github.com/spf13/pflag"
	"os"
	"strings"
)

var (
	countRepos   bool
	outputFormat string
	appConfig    *config.Config
)

// Init initializes the CLI flags and loads the configuration
func Init() {
	var err error

	// Load configuration
	appConfig, err = config.LoadConfig()
	if err != nil {
		// Non-fatal error, we'll use defaults
		fmt.Fprintf(os.Stderr, "Warning: %s\n", errors.UserFriendlyMessage(err))
		appConfig = config.DefaultConfig()
	}

	// Set up flags with defaults from config
	flag.BoolVarP(&countRepos, "count", "c", appConfig.CountRepos, "Count repos")
	flag.StringVarP(&outputFormat, "output", "o", appConfig.OutputFormat, "Output format (text, json, csv)")
}

// Parse parses the command line arguments
func Parse() []string {
	flag.Parse()
	return flag.Args()
}

// ExitWithError prints an error message and exits with code 1
func ExitWithError(err error) {
	fmt.Println(errors.UserFriendlyMessage(err))
	os.Exit(1)
}

// Run executes the CLI command
func Run(args []string) {
	target := ""

	// Use default org from config if no args provided
	if len(args) < 1 {
		if appConfig.DefaultOrg == "" {
			ExitWithError(errors.NewInputError("no GitHub organization or repository specified", nil))
		}
		target = appConfig.DefaultOrg
	} else {
		target = args[0]
	}

	// Get topics using the topics package
	topicsList, err := topics.GetTopics(target, countRepos)
	if err != nil {
		ExitWithError(err)
	}

	// Format and print topics based on output format
	switch outputFormat {
	case "json":
		printJSON(topicsList)
	case "csv":
		printCSV(topicsList)
	default:
		// Default to text format
		for _, t := range topicsList {
			fmt.Println(t)
		}
	}
}

// printJSON prints topics in JSON format
func printJSON(topics []string) {
	// Simple JSON array output
	fmt.Print("[\n")
	for i, topic := range topics {
		fmt.Printf("  %q", topic)
		if i < len(topics)-1 {
			fmt.Print(",")
		}
		fmt.Print("\n")
	}
	fmt.Print("]\n")
}

// printCSV prints topics in CSV format
func printCSV(topics []string) {
	for _, topic := range topics {
		// Handle topics with commas by quoting
		if strings.Contains(topic, ",") {
			fmt.Printf("\"%s\"\n", topic)
		} else {
			fmt.Println(topic)
		}
	}
}
