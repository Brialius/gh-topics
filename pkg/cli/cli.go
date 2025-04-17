package cli

import (
	"fmt"
	"github.com/Brialius/gh-topics/pkg/topics"
	flag "github.com/spf13/pflag"
	"os"
)

var countRepos bool

// Init initializes the CLI flags
func Init() {
	flag.BoolVarP(&countRepos, "count", "c", false, "Count repos")
}

// Parse parses the command line arguments
func Parse() []string {
	flag.Parse()
	return flag.Args()
}

// ExitWithError prints an error message and exits with code 1
func ExitWithError(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

// Run executes the CLI command
func Run(args []string) {
	if len(args) < 1 {
		ExitWithError("Please specify GitHub organization or repository (MyOrg or MyOrg/MyRepo)")
	}

	// Get topics using the topics package
	topicsList, err := topics.GetTopics(args[0], countRepos)
	if err != nil {
		ExitWithError(fmt.Sprintf("Can't get topics for %s: %s", args[0], err.Error()))
	}

	// Print topics
	for _, t := range topicsList {
		fmt.Println(t)
	}
}
