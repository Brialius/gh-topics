// Package config provides configuration management for gh-topics
package config

import (
	"fmt"
	"github.com/Brialius/gh-topics/pkg/errors"
	yaml "gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

// Config holds the application configuration
type Config struct {
	// GitHub API token
	Token string
	// Default organization to use when no organization is specified
	DefaultOrg string
	// Default output format (text, json, csv)
	OutputFormat string
	// Whether to count repositories for each topic
	CountRepos bool
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Token:        "",
		DefaultOrg:   "",
		OutputFormat: "text",
		CountRepos:   false,
	}
}

// LoadConfig loads the configuration from environment variables and config file
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// Load from config file
	configFile, err := findConfigFile()
	if err == nil {
		err = loadFromFile(config, configFile)
		if err != nil {
			return nil, errors.NewConfigError("failed to load config from file", err)
		}
	}

	// Override with environment variables
	loadFromEnv(config)

	return config, nil
}

// findConfigFile looks for a config file in standard locations
func findConfigFile() (string, error) {
	// Check for config file in current directory
	if _, err := os.Stat(".gh-topics.config"); err == nil {
		return ".gh-topics.config", nil
	}

	// Check for config file in user's home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.NewConfigError("failed to get user home directory", err)
	}

	configPath := filepath.Join(home, ".config", "gh-topics", "config")
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	// No config file found
	return "", errors.NewConfigError("no config file found", nil)
}

// loadFromFile loads configuration from a YAML file
func loadFromFile(config *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return errors.NewConfigError(fmt.Sprintf("failed to read config file %s", path), err)
	}

	// Create a temporary config to unmarshal into
	var fileConfig struct {
		Token        string `yaml:"token"`
		DefaultOrg   string `yaml:"default_org"`
		OutputFormat string `yaml:"output_format"`
		CountRepos   bool   `yaml:"count_repos"`
	}

	err = yaml.Unmarshal(data, &fileConfig)
	if err != nil {
		return errors.NewConfigError(fmt.Sprintf("failed to parse config file %s", path), err)
	}

	// Update the config with values from the file
	if fileConfig.Token != "" {
		config.Token = fileConfig.Token
	}
	if fileConfig.DefaultOrg != "" {
		config.DefaultOrg = fileConfig.DefaultOrg
	}
	if fileConfig.OutputFormat != "" {
		config.OutputFormat = fileConfig.OutputFormat
	}
	config.CountRepos = fileConfig.CountRepos

	return nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	if token := os.Getenv("GH_TOPICS_TOKEN"); token != "" {
		config.Token = token
	} else if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		config.Token = token
	}

	if org := os.Getenv("GH_TOPICS_DEFAULT_ORG"); org != "" {
		config.DefaultOrg = org
	}

	if format := os.Getenv("GH_TOPICS_OUTPUT_FORMAT"); format != "" {
		config.OutputFormat = strings.ToLower(format)
	}

	if countRepos := os.Getenv("GH_TOPICS_COUNT_REPOS"); countRepos == "true" || countRepos == "1" {
		config.CountRepos = true
	}
}
