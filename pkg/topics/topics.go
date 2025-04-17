package topics

import (
	"fmt"
	"github.com/Brialius/gh-topics/pkg/api"
	"strings"
)

// GetTopics returns a slice of topics for a repository or organization
// If countRepos is true, each topic will be prefixed with the number of repositories using it
func GetTopics(target string, countRepos bool) ([]string, error) {
	topics := make([]string, 0, 0)
	var err error

	if strings.Contains(target, "/") {
		// Target is a repository
		topics, err = api.GetRepoTopics(target)
		if err != nil {
			return nil, err
		}
	} else {
		// Target is an organization
		topicsMap, err := api.GetAllOrgTopics(target)
		if err != nil {
			return nil, err
		}

		// Format topics based on countRepos flag
		for topic, count := range topicsMap {
			if countRepos {
				topics = append(topics, formatTopicWithCount(topic, count))
			} else {
				topics = append(topics, topic)
			}
		}
	}

	return topics, nil
}

// formatTopicWithCount formats a topic with its repository count
func formatTopicWithCount(topic string, count int) string {
	return fmt.Sprintf("%d\t%s", count, topic)
}
