package topics

import (
	"strings"

	"github.com/Brialius/gh-topics/pkg/api"
)

// Topic holds a topic name and the number of repositories using it.
type Topic struct {
	Name  string `json:"topic"`
	Count int    `json:"count"`
}

// GetTopics returns topics and repository counts for a repository or organization.
// A single-repository target has a count of one for each topic.
func GetTopics(target string) ([]Topic, error) {
	result := make([]Topic, 0)
	if strings.Contains(target, "/") {
		names, err := api.GetRepoTopics(target)
		if err != nil {
			return nil, err
		}
		for _, name := range names {
			result = append(result, Topic{Name: name, Count: 1})
		}
	} else {
		counts, err := api.GetAllOrgTopics(target)
		if err != nil {
			return nil, err
		}
		for name, count := range counts {
			result = append(result, Topic{Name: name, Count: count})
		}
	}
	return result, nil
}
