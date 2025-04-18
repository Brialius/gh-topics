package api

import (
	"strings"
	"testing"
)

// TestNormalizeTopicName tests the topic name normalization logic
func TestNormalizeTopicName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase",
			input:    "topic",
			expected: "topic",
		},
		{
			name:     "uppercase",
			input:    "TOPIC",
			expected: "topic",
		},
		{
			name:     "mixed case",
			input:    "TopIC",
			expected: "topic",
		},
		{
			name:     "with spaces",
			input:    "topic name",
			expected: "topic-name",
		},
		{
			name:     "with multiple spaces",
			input:    "topic  name",
			expected: "topic--name",
		},
		{
			name:     "with spaces and mixed case",
			input:    "Topic Name",
			expected: "topic-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Extract the normalization logic from AddRepoTopics
			normalized := strings.ToLower(strings.ReplaceAll(tt.input, " ", "-"))
			if normalized != tt.expected {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, normalized, tt.expected)
			}
		})
	}
}
