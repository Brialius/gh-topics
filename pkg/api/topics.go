package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Brialius/gh-topics/pkg/errors"
	"github.com/cli/go-gh/v2/pkg/api"
	graphql "github.com/cli/shurcooL-graphql"
	"strings"
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

// AddRepoTopics adds topics to a repository
// It returns an error if the operation fails
func AddRepoTopics(repo string, newTopics []string) error {
	if len(newTopics) == 0 {
		return errors.NewInputError("no topics provided to add", nil)
	}

	// Get current topics first
	currentTopics, err := GetRepoTopics(repo)
	if err != nil {
		return errors.NewAPIError(fmt.Sprintf("failed to get current topics for repository %s", repo), err)
	}

	// Create a map for quick lookup
	topicsMap := make(map[string]bool)
	for _, topic := range currentTopics {
		topicsMap[topic] = true
	}

	// Add new topics to the map
	topicsAdded := false
	for _, topic := range newTopics {
		// Normalize topic name (lowercase, replace spaces with hyphens)
		normalizedTopic := strings.ToLower(strings.ReplaceAll(topic, " ", "-"))
		if !topicsMap[normalizedTopic] {
			topicsMap[normalizedTopic] = true
			topicsAdded = true
		}
	}

	// If no new topics were added, return early
	if !topicsAdded {
		return nil
	}

	// Convert map back to slice
	updatedTopics := make([]string, 0, len(topicsMap))
	for topic := range topicsMap {
		updatedTopics = append(updatedTopics, topic)
	}

	// Update topics via API
	return updateRepoTopics(repo, updatedTopics)
}

// RemoveRepoTopics removes topics from a repository
// It returns an error if the operation fails
func RemoveRepoTopics(repo string, topicsToRemove []string) error {
	if len(topicsToRemove) == 0 {
		return errors.NewInputError("no topics provided to remove", nil)
	}

	// Get current topics first
	currentTopics, err := GetRepoTopics(repo)
	if err != nil {
		return errors.NewAPIError(fmt.Sprintf("failed to get current topics for repository %s", repo), err)
	}

	// Create a map for quick lookup of topics to remove
	removeMap := make(map[string]bool)
	for _, topic := range topicsToRemove {
		// Normalize topic name (lowercase, replace spaces with hyphens)
		normalizedTopic := strings.ToLower(strings.ReplaceAll(topic, " ", "-"))
		removeMap[normalizedTopic] = true
	}

	// Filter out topics to remove
	updatedTopics := make([]string, 0, len(currentTopics))
	topicsRemoved := false
	for _, topic := range currentTopics {
		if !removeMap[topic] {
			updatedTopics = append(updatedTopics, topic)
		} else {
			topicsRemoved = true
		}
	}

	// If no topics were removed, return early
	if !topicsRemoved {
		return nil
	}

	// Update topics via API
	return updateRepoTopics(repo, updatedTopics)
}

// updateRepoTopics updates the topics for a repository
// It's a helper function used by both AddRepoTopics and RemoveRepoTopics
func updateRepoTopics(repo string, topics []string) error {
	opts := api.ClientOptions{
		EnableCache: false, // Don't cache for write operations
		Timeout:     GQLTimeout,
	}

	client, err := api.NewRESTClient(opts)
	if err != nil {
		return errors.NewAPIError("failed to create REST client", err)
	}

	// GitHub API expects a JSON object with a "names" array
	requestBody := struct {
		Names []string `json:"names"`
	}{
		Names: topics,
	}

	// Marshal the request body to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return errors.NewAPIError("failed to marshal request body", err)
	}

	// Create a reader from the JSON data
	bodyReader := bytes.NewReader(jsonData)

	// PUT request to update topics
	err = client.Put(fmt.Sprintf("repos/%s/topics", repo), bodyReader, nil)
	if err != nil {
		return errors.NewAPIError(fmt.Sprintf("failed to update topics for repository %s", repo), err)
	}

	return nil
}
