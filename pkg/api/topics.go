package api

import (
	"fmt"
	"github.com/Brialius/gh-topics/pkg/errors"
	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"time"
)

// GQLTimeout is the timeout for GraphQL requests
const GQLTimeout = 15 * time.Second

// GetRepoTopics returns the topics for a repo as a slice of strings or an error
func GetRepoTopics(repo string) ([]string, error) {
	opts := api.ClientOptions{
		EnableCache: true,
		Timeout:     GQLTimeout,
	}

	client, err := api.NewRESTClient(opts)
	if err != nil {
		return nil, errors.NewAPIError("failed to create REST client", err)
	}

	response := struct{ Names []string }{}

	// Assuming 100 topics per repo is enough
	err = client.Get(fmt.Sprintf("repos/%s/topics?per_page=100", repo), &response)
	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to get topics for repository %s", repo), err)
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

	client, err := api.NewGraphQLClient(opts)

	if err != nil {
		return nil, errors.NewAPIError(fmt.Sprintf("failed to create GraphQL client for organization %s", org), err)
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
			return nil, errors.NewAPIError(fmt.Sprintf("GraphQL query failed for organization %s", org), err)
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
