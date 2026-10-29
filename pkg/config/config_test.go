package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	got := DefaultConfig()
	if *got != (Config{OutputFormat: "text"}) {
		t.Fatalf("unexpected defaults: %+v", got)
	}
}

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name, file, envOrg, envFormat, envCount string
		directory, absent                       bool
		want                                    Config
		errorText                               string
	}{
		{name: "absent", absent: true, want: Config{OutputFormat: "text"}},
		{name: "empty", want: Config{OutputFormat: "text"}},
		{name: "file settings", file: "default_org: file-org\noutput_format: JSON\ncount_repos: true\n", want: Config{DefaultOrg: "file-org", OutputFormat: "json", CountRepos: true}},
		{name: "environment overrides file", file: "default_org: file-org\noutput_format: json\n", envOrg: "env-org", envFormat: "CSV", envCount: "true", want: Config{DefaultOrg: "env-org", OutputFormat: "csv", CountRepos: true}},
		{name: "numeric true", envCount: "1", want: Config{OutputFormat: "text", CountRepos: true}},
		{name: "malformed file preserves environment", file: "[", envOrg: "env-org", envFormat: "JSON", envCount: "true", want: Config{DefaultOrg: "env-org", OutputFormat: "json", CountRepos: true}, errorText: "failed to parse config file"},
		{name: "unreadable file preserves environment", directory: true, envOrg: "env-org", envFormat: "CSV", envCount: "1", want: Config{DefaultOrg: "env-org", OutputFormat: "csv", CountRepos: true}, errorText: "failed to read config file"},
		{name: "partial malformed values are discarded", file: "default_org: partial\ncount_repos: [bad]\n", want: Config{OutputFormat: "text"}, errorText: "failed to parse config file"},
		{name: "unsupported token ignored", file: "token: unused\n", want: Config{OutputFormat: "text"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := isolatedConfig(t)
			t.Setenv("GH_TOPICS_DEFAULT_ORG", tt.envOrg)
			t.Setenv("GH_TOPICS_OUTPUT_FORMAT", tt.envFormat)
			t.Setenv("GH_TOPICS_COUNT_REPOS", tt.envCount)
			path := filepath.Join(dir, ".gh-topics.config")
			if tt.directory {
				if err := os.Mkdir(path, 0700); err != nil {
					t.Fatal(err)
				}
			} else if !tt.absent {
				if err := os.WriteFile(path, []byte(tt.file), 0600); err != nil {
					t.Fatal(err)
				}
			}
			got, err := LoadConfig()
			if tt.errorText == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.errorText) || !strings.Contains(err.Error(), ".gh-topics.config") {
				t.Fatalf("error = %v, want file path and %q", err, tt.errorText)
			}
			if got == nil || *got != tt.want {
				t.Fatalf("config = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestConfigFilePrecedence(t *testing.T) {
	dir := isolatedConfig(t)
	homePath := filepath.Join(dir, ".config", "gh-topics")
	if err := os.MkdirAll(homePath, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(homePath, "config"), []byte("default_org: home-org\noutput_format: csv\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := LoadConfig()
	if err != nil || got.DefaultOrg != "home-org" {
		t.Fatalf("home config = %+v, %v", got, err)
	}
	if err := os.WriteFile(".gh-topics.config", []byte("default_org: local-org\n"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err = LoadConfig()
	if err != nil || got.DefaultOrg != "local-org" || got.OutputFormat != "text" {
		t.Fatalf("local config = %+v, %v", got, err)
	}
}

func isolatedConfig(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Error(err)
		}
	})
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	t.Setenv("home", dir)
	for _, key := range []string{"GH_TOPICS_DEFAULT_ORG", "GH_TOPICS_OUTPUT_FORMAT", "GH_TOPICS_COUNT_REPOS"} {
		t.Setenv(key, "")
	}
	return dir
}

func TestMissingHomeIsOptional(t *testing.T) {
	isolatedConfig(t)
	for _, key := range []string{"HOME", "USERPROFILE", "home"} {
		t.Setenv(key, "")
	}
	t.Setenv("GH_TOPICS_DEFAULT_ORG", "env-org")
	got, err := LoadConfig()
	if err != nil || got.DefaultOrg != "env-org" {
		t.Fatalf("config = %+v, error = %v", got, err)
	}
}
