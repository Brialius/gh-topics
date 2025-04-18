# gh-topics

[GitHub CLI](https://github.com/cli/cli) extension to work with topics

[![Go Report Card](https://goreportcard.com/badge/github.com/Brialius/gh-topics)](https://goreportcard.com/report/github.com/Brialius/gh-topics)

## Installation
```bash
gh extension install Brialius/gh-topics
```

## Usage
```bash
gh topics [options] [organization|repository]
```

If no organization or repository is specified, the default organization from the configuration will be used.

### Options

- `-c, --count`: Count repositories for each topic (only applicable for organizations)
- `-o, --output`: Output format (text, json, csv)
- `-h, --help`: Show help message

### Examples

```bash
# Get topics for a repository
gh topics Brialius/gh-topics

# Get topics for an organization with repository counts
gh topics Brialius --count

# Get topics for an organization in JSON format
gh topics Brialius --output json

# Get topics for an organization in CSV format
gh topics Brialius --output csv
```

## Configuration

gh-topics can be configured using environment variables or a configuration file.

### Environment Variables

- `GH_TOPICS_TOKEN`: GitHub API token (falls back to `GITHUB_TOKEN`)
- `GH_TOPICS_DEFAULT_ORG`: Default organization to use when no organization is specified
- `GH_TOPICS_OUTPUT_FORMAT`: Default output format (text, json, csv)
- `GH_TOPICS_COUNT_REPOS`: Whether to count repositories for each topic (true/false)

### Configuration File

gh-topics looks for a configuration file in the following locations:
1. `.gh-topics.config` in the current directory
2. `~/.config/gh-topics/config`

The configuration file should be in YAML format:

```yaml
# GitHub API token
token: ghp_your_token_here
# Default organization
default_org: YourOrg
# Default output format (text, json, csv)
output_format: text
# Whether to count repositories for each topic
count_repos: false
```

## Error Handling

gh-topics provides user-friendly error messages for different types of errors:

- API errors: Issues with the GitHub API
- Configuration errors: Problems with the configuration
- Input errors: Invalid command-line arguments
- Internal errors: Unexpected application errors
