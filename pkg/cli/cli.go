package cli

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/Brialius/gh-topics/pkg/config"
	"github.com/Brialius/gh-topics/pkg/errors"
	"github.com/Brialius/gh-topics/pkg/topics"
	"github.com/muesli/termenv"
	flag "github.com/spf13/pflag"
	"io"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
)

var (
	Version      = "dev"
	showHelp     bool
	showVersion  bool
	countRepos   bool
	outputFormat string
	appConfig    *config.Config

	// Terminal output
	term        = termenv.EnvColorProfile()
	colorOutput = true
)

// Init initializes the CLI flags. Informational commands do not load configuration.
func Init() {
	flag.BoolVarP(&countRepos, "count", "c", false, "Count repositories per topic (organizations only)")
	flag.StringVarP(&outputFormat, "output", "o", "text", "Output format (text, json, csv)")
	flag.BoolVar(&colorOutput, "color", true, "Enable colorized output")
	flag.BoolVarP(&showHelp, "help", "h", false, "Show help")
	flag.BoolVarP(&showVersion, "version", "v", false, "Show version")
	flag.Usage = func() { printHelp(os.Stderr) }
}

// versionString identifies release builds and local builds with VCS metadata.
func versionString() string {
	if Version != "" && Version != "dev" {
		return Version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	return developmentVersion(info)
}

func developmentVersion(info *debug.BuildInfo) string {
	revision := ""
	dirty := false
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			dirty = setting.Value == "true"
		}
	}
	if revision == "" {
		return "dev"
	}
	if len(revision) > 7 {
		revision = revision[:7]
	}
	if dirty {
		revision += "-dirty"
	}
	return "dev (" + revision + ")"
}

func printHelp(out io.Writer) {
	_, _ = fmt.Fprintf(out, `gh-topics version %s

List repository topics and discover topics across an organization.

USAGE
  gh topics [options] [ORGANIZATION | OWNER/REPOSITORY]

OPTIONS
%s
EXAMPLES
  gh topics Brialius/gh-topics
  gh topics cli --count
  gh topics cli --output json

DEFAULT TARGET
  With no target, use default_org or GH_TOPICS_DEFAULT_ORG.
  If neither is set, show this help. The current repository is not selected automatically.
  Configuration can change the defaults for --count and --output.
`, versionString(), flag.CommandLine.FlagUsages())
}

// Parse parses the command line arguments
func Parse() []string {
	flag.Parse()
	return flag.Args()
}

// ExitWithError prints an error message and exits with code 1
func ExitWithError(err error) {
	fmt.Fprintln(os.Stderr, colorError(errors.UserFriendlyMessage(err)))
	os.Exit(1)
}

// Run executes the CLI command
func Run(args []string) {
	if showHelp {
		printHelp(os.Stdout)
		return
	}
	if showVersion {
		fmt.Printf("gh-topics version %s\n", versionString())
		return
	}
	if len(args) > 1 {
		ExitWithError(errors.NewInputError("expected one organization or repository", nil))
	}
	var err error
	appConfig, err = config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", errors.UserFriendlyMessage(err))
	}
	if !flag.CommandLine.Changed("count") {
		countRepos = appConfig.CountRepos
	}
	if !flag.CommandLine.Changed("output") {
		outputFormat = appConfig.OutputFormat
	}
	target := ""

	// Use default org from config if no args provided
	if len(args) < 1 {
		if appConfig.DefaultOrg == "" {
			printHelp(os.Stdout)
			return
		}
		target = appConfig.DefaultOrg
	} else {
		target = args[0]
	}

	// Get topics using the topics package
	topicsList, err := topics.GetTopics(target)
	if err != nil {
		ExitWithError(err)
	}

	includeCounts := countRepos && !strings.Contains(target, "/")
	if err := writeTopics(os.Stdout, topicsList, outputFormat, includeCounts); err != nil {
		ExitWithError(errors.NewInternalError("failed to write topics", err))
	}
}

// writeTopics encodes structured data without putting presentation delimiters in fields.
func writeTopics(out io.Writer, list []topics.Topic, format string, includeCounts bool) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if includeCounts {
			if list == nil {
				list = []topics.Topic{}
			}
			return encoder.Encode(list)
		}
		names := make([]string, 0, len(list))
		for _, topic := range list {
			names = append(names, topic.Name)
		}
		return encoder.Encode(names)
	case "csv":
		writer := csv.NewWriter(out)
		if includeCounts {
			if err := writer.Write([]string{"topic", "count"}); err != nil {
				return err
			}
		}
		for _, topic := range list {
			row := []string{topic.Name}
			if includeCounts {
				row = append(row, strconv.Itoa(topic.Count))
			}
			if err := writer.Write(row); err != nil {
				return err
			}
		}
		writer.Flush()
		return writer.Error()
	default:
		for _, topic := range list {
			line := colorTopic(topic.Name)
			if includeCounts {
				line = colorCount(strconv.Itoa(topic.Count)) + "\t" + line
			}
			if _, err := fmt.Fprintln(out, line); err != nil {
				return err
			}
		}
		return nil
	}
}

// colorError returns a string formatted as an error message
func colorError(msg string) string {
	if !colorOutput {
		return msg
	}
	return termenv.String(msg).Foreground(term.Color("1")).String() // Red
}

// colorTopic returns a string formatted as a topic
func colorTopic(topic string) string {
	if !colorOutput {
		return topic
	}
	return termenv.String(topic).Foreground(term.Color("4")).String() // Blue
}

// colorCount returns a string formatted as a count
func colorCount(count string) string {
	if !colorOutput {
		return count
	}
	return termenv.String(count).Foreground(term.Color("3")).String() // Yellow
}
