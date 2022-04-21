package main

import (
	"fmt"
	"github.com/cli/go-gh/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	flag "github.com/spf13/pflag"
	"os"
	"strings"
	"time"

	"github.com/cli/go-gh"
)

// GQLTimeout is the timeout for GraphQL requests
const GQLTimeout = 15 * time.Second

var countRepos bool

func init() {
	flag.BoolVarP(&countRepos, "count", "c", false, "Count repos")
}

// GetRepoTopics returns the topics for a repo as a slice of strings or an error
func GetRepoTopics(repo string) ([]string, error) {
	opts := api.ClientOptions{
		EnableCache: true,
		Timeout:     GQLTimeout,
	}

	client, err := gh.RESTClient(&opts)
	if err != nil {
		return nil, err
	}

	response := struct{ Names []string }{}

	// Assuming 100 topics per repo is enough
	err = client.Get(fmt.Sprintf("repos/%s/topics?per_page=100", repo), &response)
	if err != nil {
		return nil, err
	}

	return response.Names, nil
}

// GetAllOrgTopics returns a map with all topics for an organization or error
// keys are topic names, values are repo counts
func GetAllOrgTopics(org string) (map[string]int, error) {
	opts := api.ClientOptions{
		EnableCache: true,
		Timeout:     GQLTimeout,
	}

	client, err := gh.GQLClient(&opts)

	if err != nil {
		return nil, fmt.Errorf("can't connect to Github: %w", err)
	}

	var query struct {
		Organization struct {
			Repositories struct {
				PageInfo struct {
					EndCursor   graphql.String
					HasNextPage bool
				}
				Nodes []struct {
					RepositoryTopics struct {
						Nodes []struct {
							Topic struct {
								Name string
							}
						}
					} `graphql:"repositoryTopics(first: 100)"` // Assuming 100 topics per repo is enough
				}
			} `graphql:"repositories(first: $first, after: $after)"`
		} `graphql:"organization(login: $login)"`
	}

	var after *graphql.String
	topicsMap := map[string]int{}

	for {
		variables := map[string]interface{}{
			"first": graphql.Int(100),
			"login": graphql.String(org),
			"after": after,
		}

		err = client.Query("Topics", &query, variables)

		if err != nil {
			return nil, fmt.Errorf("GQL request query was failed: %w", err)
		}

		after = &query.Organization.Repositories.PageInfo.EndCursor

		for _, r := range query.Organization.Repositories.Nodes {
			for _, t := range r.RepositoryTopics.Nodes {
				if t.Topic.Name != "" {
					topicsMap[t.Topic.Name] = topicsMap[t.Topic.Name] + 1
				}
			}
		}

		if !query.Organization.Repositories.PageInfo.HasNextPage {
			break
		}
	}

	return topicsMap, nil
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
		topics, err = GetRepoTopics(args[0])
		if err != nil {
			exitWithError(fmt.Sprintf("Can't get topics for %s: %s", args[0], err.Error()))
		}
	} else {
		topicsMap, err := GetAllOrgTopics(args[0])
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
