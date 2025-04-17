package main

import (
	"fmt"
	"github.com/Brialius/gh-topics/pkg/api"
	flag "github.com/spf13/pflag"
	"os"
	"strings"
)

var countRepos bool

func init() {
	flag.BoolVarP(&countRepos, "count", "c", false, "Count repos")
}

func exitWithError(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func main() {
	var err error

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		exitWithError("Please specify GitHub organization or repository (MyOrg	or MyOrg/MyRepo)")
	}

	topics := make([]string, 0, 0)

	if strings.Contains(args[0], "/") {
		topics, err = api.GetRepoTopics(args[0])
		if err != nil {
			exitWithError(fmt.Sprintf("Can't get topics for %s: %s", args[0], err.Error()))
		}
	} else {
		topicsMap, err := api.GetAllOrgTopics(args[0])
		if err != nil {
			exitWithError(fmt.Sprintf("Can't get topics for %s: %s", args[0], err.Error()))
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
