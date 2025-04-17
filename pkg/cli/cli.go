package cli

import (
	"fmt"
	"github.com/Brialius/gh-topics/pkg/api"
	flag "github.com/spf13/pflag"
	"os"
	"strings"
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

	topics := make([]string, 0, 0)
	var err error

	if strings.Contains(args[0], "/") {
		topics, err = api.GetRepoTopics(args[0])
		if err != nil {
			ExitWithError(fmt.Sprintf("Can't get topics for %s: %s", args[0], err.Error()))
		}
	} else {
		topicsMap, err := api.GetAllOrgTopics(args[0])
		if err != nil {
			ExitWithError(fmt.Sprintf("Can't get topics for %s: %s", args[0], err.Error()))
		}

		for t, c := range topicsMap {
			if countRepos {
				topics = append(topics, fmt.Sprintf("%d\t%s", c, t))
			} else {
				topics = append(topics, t)
			}
		}
	}

	for _, t := range topics {
		fmt.Println(t)
	}
}
