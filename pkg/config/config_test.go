package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Token != "" {
		t.Errorf("DefaultConfig().Token = %v, want %v", config.Token, "")
	}
	if config.DefaultOrg != "" {
		t.Errorf("DefaultConfig().DefaultOrg = %v, want %v", config.DefaultOrg, "")
	}
	if config.OutputFormat != "text" {
		t.Errorf("DefaultConfig().OutputFormat = %v, want %v", config.OutputFormat, "text")
	}
	if config.CountRepos != false {
		t.Errorf("DefaultConfig().CountRepos = %v, want %v", config.CountRepos, false)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Save original environment variables
	origToken := os.Getenv("GH_TOPICS_TOKEN")
	origOrg := os.Getenv("GH_TOPICS_DEFAULT_ORG")
	origFormat := os.Getenv("GH_TOPICS_OUTPUT_FORMAT")
	origCount := os.Getenv("GH_TOPICS_COUNT_REPOS")

	// Restore environment variables after test
	defer func() {
		os.Setenv("GH_TOPICS_TOKEN", origToken)
		os.Setenv("GH_TOPICS_DEFAULT_ORG", origOrg)
		os.Setenv("GH_TOPICS_OUTPUT_FORMAT", origFormat)
		os.Setenv("GH_TOPICS_COUNT_REPOS", origCount)
	}()

	// Set test environment variables
	os.Setenv("GH_TOPICS_TOKEN", "test-token")
	os.Setenv("GH_TOPICS_DEFAULT_ORG", "test-org")
	os.Setenv("GH_TOPICS_OUTPUT_FORMAT", "json")
	os.Setenv("GH_TOPICS_COUNT_REPOS", "true")

	config := DefaultConfig()
	loadFromEnv(config)

	if config.Token != "test-token" {
		t.Errorf("loadFromEnv().Token = %v, want %v", config.Token, "test-token")
	}
	if config.DefaultOrg != "test-org" {
		t.Errorf("loadFromEnv().DefaultOrg = %v, want %v", config.DefaultOrg, "test-org")
	}
	if config.OutputFormat != "json" {
		t.Errorf("loadFromEnv().OutputFormat = %v, want %v", config.OutputFormat, "json")
	}
	if config.CountRepos != true {
		t.Errorf("loadFromEnv().CountRepos = %v, want %v", config.CountRepos, true)
	}
}

func TestLoadFromEnv_GithubToken(t *testing.T) {
	// Save original environment variables
	origToken := os.Getenv("GH_TOPICS_TOKEN")
	origGithubToken := os.Getenv("GITHUB_TOKEN")

	// Restore environment variables after test
	defer func() {
		os.Setenv("GH_TOPICS_TOKEN", origToken)
		os.Setenv("GITHUB_TOKEN", origGithubToken)
	}()

	// Unset GH_TOPICS_TOKEN and set GITHUB_TOKEN
	os.Unsetenv("GH_TOPICS_TOKEN")
	os.Setenv("GITHUB_TOKEN", "github-token")

	config := DefaultConfig()
	loadFromEnv(config)

	if config.Token != "github-token" {
		t.Errorf("loadFromEnv().Token = %v, want %v", config.Token, "github-token")
	}
}

func TestLoadFromEnv_CountReposNumeric(t *testing.T) {
	// Save original environment variable
	origCount := os.Getenv("GH_TOPICS_COUNT_REPOS")

	// Restore environment variable after test
	defer func() {
		os.Setenv("GH_TOPICS_COUNT_REPOS", origCount)
	}()

	// Set test environment variable
	os.Setenv("GH_TOPICS_COUNT_REPOS", "1")

	config := DefaultConfig()
	loadFromEnv(config)

	if config.CountRepos != true {
		t.Errorf("loadFromEnv().CountRepos = %v, want %v", config.CountRepos, true)
	}
}
