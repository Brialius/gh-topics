package cli

import (
	"bytes"
	"encoding/json"
	appErrors "github.com/Brialius/gh-topics/pkg/errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	flag "github.com/spf13/pflag"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestCLI(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		config string
		org    string
		want   string
		target string
	}{
		{name: "help ignores invalid config", args: []string{"--help"}, config: "[", want: "USAGE"},
		{name: "short help", args: []string{"-h"}, want: "EXAMPLES"},
		{name: "version ignores invalid config", args: []string{"--version"}, config: "[", want: "gh-topics version v1.2.3\n"},
		{name: "short version", args: []string{"-v"}, want: "gh-topics version v1.2.3\n"},
		{name: "no target", want: "DEFAULT TARGET"},
		{name: "invalid file retains env target", config: "[", org: "env-org", args: []string{"--output=json"}, want: "[\n  \"go\"\n]\n", target: "env-org"},
		{name: "uppercase file format", config: "default_org: file-org\noutput_format: JSON\n", want: "[\n  \"go\"\n]\n", target: "file-org"},
		{name: "configured organization", config: "default_org: file-org\ncount_repos: true\noutput_format: json\n", want: "[\n  {\n    \"topic\": \"go\",\n    \"count\": 1\n  }\n]\n", target: "file-org"},
		{name: "organization CSV counts", args: []string{"cli", "--count", "--output=csv"}, want: "topic,count\ngo,1\n", target: "cli"},
		{name: "repository count ignored", args: []string{"owner/repo", "--count", "--output=json"}, want: "[\n  \"go\"\n]\n", target: "owner/repo"},
		{name: "environment overrides file", config: "default_org: file-org\n", org: "env-org", args: []string{"--color=false"}, want: "go\n", target: "env-org"},
		{name: "explicit flags override file", config: "default_org: file-org\ncount_repos: true\noutput_format: json\n", args: []string{"cli", "--count=false", "--output=csv"}, want: "go\n", target: "cli"},
		{name: "repository JSON has no banner", args: []string{"owner/repo", "--output=json"}, want: "[\n  \"go\"\n]\n", target: "owner/repo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
			for _, key := range []string{"GH_TOPICS_DEFAULT_ORG", "GH_TOPICS_OUTPUT_FORMAT", "GH_TOPICS_COUNT_REPOS", "GH_DEBUG"} {
				t.Setenv(key, "")
			}
			t.Setenv("HOME", dir)
			t.Setenv("USERPROFILE", dir)
			t.Setenv("home", dir)
			t.Setenv("GH_CONFIG_DIR", dir)
			t.Setenv("GH_CACHE_DIR", filepath.Join(dir, "cache"))
			t.Setenv("GH_HOST", "github.com")
			t.Setenv("GH_TOKEN", "test-token")
			t.Setenv("GH_TOPICS_DEFAULT_ORG", tt.org)
			if tt.config != "" {
				if err := os.WriteFile(".gh-topics.config", []byte(tt.config), 0600); err != nil {
					t.Fatal(err)
				}
			}
			oldFlags, oldVersion, oldTransport := flag.CommandLine, Version, http.DefaultTransport
			t.Cleanup(func() { flag.CommandLine, Version, http.DefaultTransport = oldFlags, oldVersion, oldTransport })
			flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
			Version = "v1.2.3"
			calls := 0
			http.DefaultTransport = transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if tt.target == "" {
					t.Fatal("informational command made an API request")
				}
				body := `{"names":["go"]}`
				if strings.Contains(tt.target, "/") {
					if r.URL.Path != "/repos/"+tt.target+"/topics" {
						t.Errorf("unexpected path: %s", r.URL.Path)
					}
				} else {
					var query struct{ Variables struct{ Login string } }
					if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
						t.Fatal(err)
					}
					if query.Variables.Login != tt.target {
						t.Errorf("target = %q, want %q", query.Variables.Login, tt.target)
					}
					body = `{"data":{"organization":{"repositories":{"pageInfo":{"endCursor":"","hasNextPage":false},"nodes":[{"repositoryTopics":{"nodes":[{"topic":{"name":"go"}}]}}]}}}}`
				}
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
			})
			out, err := os.CreateTemp(dir, "stdout")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := out.Close(); err != nil {
					t.Error(err)
				}
			}()
			oldOut := os.Stdout
			os.Stdout = out
			t.Cleanup(func() { os.Stdout = oldOut })
			Init()
			if err := flag.CommandLine.Parse(tt.args); err != nil {
				t.Fatal(err)
			}
			Run(flag.Args())
			data, err := os.ReadFile(out.Name())
			if err != nil {
				t.Fatal(err)
			}
			got := string(data)
			if tt.target != "" || strings.Contains(tt.name, "version") {
				if got != tt.want {
					t.Errorf("output = %q, want %q", got, tt.want)
				}
			} else if !strings.Contains(got, tt.want) {
				t.Errorf("output missing %q: %s", tt.want, got)
			}
			if tt.target != "" && calls != 1 {
				t.Errorf("API calls = %d, want 1", calls)
			}
		})
	}
}

func TestDevelopmentVersion(t *testing.T) {
	for _, tt := range []struct{ revision, modified, want string }{
		{"", "", "dev"},
		{"1234567890", "false", "dev (1234567)"},
		{"1234567890", "true", "dev (1234567-dirty)"},
		{"abc", "false", "dev (abc)"},
	} {
		info := &debug.BuildInfo{Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: tt.revision}, {Key: "vcs.modified", Value: tt.modified}}}
		if got := developmentVersion(info); got != tt.want {
			t.Errorf("version = %q, want %q", got, tt.want)
		}
	}
}

func TestFatalErrorUsesStderr(t *testing.T) {
	if os.Getenv("GH_TOPICS_TEST_FATAL") == "1" {
		colorOutput = false
		ExitWithError(appErrors.NewInputError("invalid target", nil))
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestFatalErrorUsesStderr$")
	cmd.Env = append(os.Environ(), "GH_TOPICS_TEST_FATAL=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	if err == nil || cmd.ProcessState.ExitCode() != 1 {
		t.Fatalf("exit = %v, want 1", err)
	}
	if stdout.Len() != 0 || !strings.Contains(stderr.String(), "invalid target") {
		t.Fatalf("stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}
