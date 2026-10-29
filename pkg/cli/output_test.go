package cli

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/Brialius/gh-topics/pkg/topics"
)

func TestStructuredCounts(t *testing.T) {
	list := []topics.Topic{{Name: "nlp", Count: 6}, {Name: "pam", Count: 2}}
	var out bytes.Buffer
	if err := writeTopics(&out, list, "json", true); err != nil {
		t.Fatal(err)
	}
	var decoded []struct {
		Topic string `json:"topic"`
		Count int    `json:"count"`
	}
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 2 || decoded[0].Topic != "nlp" || decoded[0].Count != 6 || decoded[1].Count != 2 {
		t.Fatalf("unexpected JSON: %s", out.String())
	}
	out.Reset()
	if err := writeTopics(&out, list, "csv", true); err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(&out).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"topic", "count"}, {"nlp", "6"}, {"pam", "2"}}
	if !reflect.DeepEqual(rows, want) {
		t.Fatalf("CSV = %v, want %v", rows, want)
	}
}

func TestOutputBoundaries(t *testing.T) {
	oldColor := colorOutput
	colorOutput = false
	t.Cleanup(func() { colorOutput = oldColor })
	for _, tt := range []struct {
		name, format string
		counts       bool
		list         []topics.Topic
		want         string
	}{
		{name: "empty JSON counts", format: "json", counts: true, want: "[]\n"},
		{name: "empty JSON names", format: "json", want: "[]\n"},
		{name: "empty CSV counts retains schema", format: "csv", counts: true, want: "topic,count\n"},
		{name: "empty CSV names", format: "csv", want: ""},
		{name: "text counts unchanged", format: "text", counts: true, list: []topics.Topic{{Name: "nlp", Count: 6}}, want: "6\tnlp\n"},
		{name: "text names unchanged", format: "text", list: []topics.Topic{{Name: "nlp", Count: 6}}, want: "nlp\n"},
		{name: "CSV quotes", format: "csv", counts: true, list: []topics.Topic{{Name: "a,\"b", Count: 6}}, want: "topic,count\n\"a,\"\"b\",6\n"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			if err := writeTopics(&out, tt.list, tt.format, tt.counts); err != nil {
				t.Fatal(err)
			}
			if out.String() != tt.want {
				t.Fatalf("output = %q, want %q", out.String(), tt.want)
			}
		})
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func TestOutputWriteErrors(t *testing.T) {
	for _, format := range []string{"json", "csv", "text"} {
		for _, counts := range []bool{false, true} {
			err := writeTopics(failingWriter{}, []topics.Topic{{Name: strings.Repeat("x", 5000), Count: 6}}, format, counts)
			if !errors.Is(err, io.ErrClosedPipe) {
				t.Errorf("%s counts=%v: %v", format, counts, err)
			}
		}
	}
}
