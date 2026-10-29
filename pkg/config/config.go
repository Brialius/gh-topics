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
		DefaultOrg:   "",
		OutputFormat: "text",
		CountRepos:   false,
	}
}

// LoadConfig loads the configuration from environment variables and config file
func LoadConfig() (*Config, error) {
	config := DefaultConfig()

	// A bad file must not prevent higher-priority environment settings from loading.
	configFile, err := findConfigFile()
	if err == nil && configFile != "" {
		err = loadFromFile(config, configFile)
	}
	loadFromEnv(config)
	return config, err
}

// findConfigFile returns the first existing file, or an empty path if none exists.
func findConfigFile() (string, error) {
	local := ".gh-topics.config"
	if _, err := os.Stat(local); err == nil {
		return local, nil
	} else if !os.IsNotExist(err) {
		return "", errors.NewConfigError("failed to inspect "+local, err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.NewConfigError("failed to get user home directory", err)
	}
	path := filepath.Join(home, ".config", "gh-topics", "config")
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if !os.IsNotExist(err) {
		return "", errors.NewConfigError("failed to inspect "+path, err)
	}
	return "", nil
}

// loadFromFile loads configuration from a YAML file
func loadFromFile(config *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return errors.NewConfigError(fmt.Sprintf("failed to read config file %s", path), err)
	}

	// Create a temporary config to unmarshal into
	var fileConfig struct {
		DefaultOrg   string `yaml:"default_org"`
		OutputFormat string `yaml:"output_format"`
		CountRepos   bool   `yaml:"count_repos"`
	}

	err = yaml.Unmarshal(data, &fileConfig)
	if err != nil {
		return errors.NewConfigError(fmt.Sprintf("failed to parse config file %s", path), err)
	}

	// Update the config with values from the file
	if fileConfig.DefaultOrg != "" {
		config.DefaultOrg = fileConfig.DefaultOrg
	}
	if fileConfig.OutputFormat != "" {
		config.OutputFormat = strings.ToLower(fileConfig.OutputFormat)
	}
	config.CountRepos = fileConfig.CountRepos

	return nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
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
