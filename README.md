# gh-topics

[GitHub CLI](https://github.com/cli/cli) extension for listing repository topics and discovering topics across an organization's repositories. Count how many repositories use each topic, or export the results as JSON or CSV.

**Built for fast organization-wide discovery:** fetch topics for up to 100 repositories in one GraphQL request, with repository counts computed locally.

## Performance

For an organization, `gh-topics` fetches repositories and their topics together in pages of 100. It aggregates topics and counts in memory, so `--count` adds no API requests. A single-repository lookup uses one REST request.

Compared with a script that calls the [repository topics endpoint](https://docs.github.com/en/rest/repos/repos#get-all-repository-topics) once for every repository:

| Repositories in an organization | One topics request per repository | `gh-topics` GraphQL requests |
| --- | ---: | ---: |
| 100 | 100 | 1 |
| 500 | 500 | 5 |
| 1,000 | 1,000 | 10 |

These are calculated request counts for successful, uncached queries, not measured timings. The per-repository approach assumes repository names are already known; discovering them requires additional requests. Fewer network round trips can make a substantial difference on large organizations, especially over high-latency connections.

`gh-topics` uses GitHub's API itself: a custom script using the same batched GraphQL query can achieve the same request count. Actual runtime depends on network latency, GitHub response times, and caching; the table does not imply a proportional speedup or equivalent API rate-limit costs.

## Installation

Install GitHub CLI (`gh`), authenticate, and install the extension:

```bash
gh auth login
gh extension install Brialius/gh-topics
```

To update an existing installation:

```bash
gh extension upgrade gh-topics
```

## Usage

```bash
gh topics [options] [ORGANIZATION | OWNER/REPOSITORY]
```

- Use `OWNER/REPOSITORY` to list a single repository's topics.
- Use an organization login to list unique topics across its repositories. Personal account logins are not supported as organization targets.
- Omit the target to use `default_org` from configuration. If no default is set, the CLI shows help and exits successfully. The current Git repository is not selected automatically.

Results include only repositories accessible to the authenticated account. Organization results are not sorted.

### Options

| Option | Description | Default |
| --- | --- | --- |
| `-c, --count` | Prefix each topic with its repository count (organizations only). | `false` |
| `-o, --output FORMAT` | Output format: `text`, `json`, or `csv`. | `text` |
| `--color` | Enable colorized text output; use `--color=false` to disable it. | `true` |
| `-h, --help` | Show version, usage, options, and examples. | — |
| `-v, --version` | Show the installed version. | — |

`--help` and `--version` exit successfully without loading configuration or making API requests. Ordinary topic output does not include a version banner.

Configuration can change the defaults for `--count` and `--output`. Use `--count=false` to disable a configured count setting.

The CLI currently lists topics only. `--add` and `--remove` are not supported.

### Examples

```bash
# List topics for a repository
gh topics Brialius/gh-topics

# List unique topics across an organization
gh topics cli

# Count repositories using each topic
gh topics cli --count

# Export organization topics
gh topics cli --output json
gh topics cli --output csv

# Sort topics alphabetically without color codes
gh topics cli --color=false | sort

# Show the most widely used topics first
gh topics cli --count --color=false | sort -nr

# Use a configured default organization
GH_TOPICS_DEFAULT_ORG=cli gh topics
```

### Output formats

- **Text:** one topic per line. With `--count`, each line contains the count, a tab, and the topic name.
- **JSON:** an array of topic-name strings. With `--count`, an array of objects with a `topic` string and a numeric `count`, for example `[{"topic":"cli","count":3},{"topic":"golang","count":2}]`.
- **CSV:** one topic per row, with no header. With `--count`, a `topic,count` header followed by separate topic and count columns, for example `cli,3`.

The counts above are illustrative. `--count` has no effect on a single-repository query. Unrecognized output formats fall back to text.

## Authentication

The extension uses GitHub CLI authentication. For interactive use, sign in with `gh auth login`. For automation against github.com, provide `GH_TOKEN` or `GITHUB_TOKEN`; `GH_TOKEN` takes precedence.

The `GH_TOPICS_TOKEN` environment variable and YAML `token` field are not supported. Use GitHub CLI authentication or the standard token environment variables instead.

## Configuration

Settings are loaded from a YAML file, then overridden by environment variables, then by explicit CLI flags where available.

### Configuration file

The first existing file is used; the two files are not merged:

1. `.gh-topics.config` in the current directory.
2. `~/.config/gh-topics/config`.

```yaml
default_org: cli
output_format: text
count_repos: false
```

A configuration file is optional. If the selected file cannot be read or parsed, the CLI prints a warning and uses built-in defaults for file settings, then applies environment variables and explicit CLI flags.

### Environment variables

| Variable | Description | Default |
| --- | --- | --- |
| `GH_TOPICS_DEFAULT_ORG` | Organization to query when no target is supplied. | None |
| `GH_TOPICS_OUTPUT_FORMAT` | Output format: `text`, `json`, or `csv`. | `text` |
| `GH_TOPICS_COUNT_REPOS` | Set to `true` or `1` to enable repository counts. | `false` |

`GH_TOPICS_COUNT_REPOS=false` and `0` do not override `count_repos: true` from a file. Use `--count=false` to disable counts explicitly.

## Troubleshooting

- **Help appears instead of topics:** provide a target or set `GH_TOPICS_DEFAULT_ORG` / `default_org`.
- **Authentication or access errors:** run `gh auth status` and check that the authenticated account can access the target repositories.
- **Organization lookup fails:** use an organization login, or `OWNER/REPOSITORY` for a repository owned by a personal account.
- **Configuration warning:** check the YAML syntax and readability of the selected configuration file.

## Development

Go 1.22 or later is required to build from source.

```bash
git clone https://github.com/Brialius/gh-topics.git
cd gh-topics
go build -o gh-topics .

# Run the local binary
./gh-topics Brialius/gh-topics

# Run tests with race detection
go test -race ./...
```

To use the local build through `gh`, run `gh extension install .` after building. If a released copy is already installed, remove it first with `gh extension remove topics`. Rebuild the binary after code changes.

Local builds report `dev` with a short commit SHA and a `dirty` suffix when VCS metadata is available. Release builds report the release version. To set a version explicitly, use `make build VERSION=v1.2.3`.

Run `make help` for additional build, test, and lint targets.

## License

[MIT](LICENSE).
