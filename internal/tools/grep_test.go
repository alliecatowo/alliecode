package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestBuildRipgrepArgsUsesIncludeOverGlob(t *testing.T) {
	in := grepInput{
		Pattern: "foo",
		Glob:    "*.go",
		Include: "*.ts",
	}
	args := buildRipgrepArgs(in, "files_with_matches", "/tmp")

	expected := []string{"--sort", "path", "--files-with-matches", "--glob", "*.ts", "foo", "/tmp"}
	if !reflect.DeepEqual(args, expected) {
		t.Fatalf("unexpected args:\nwant: %#v\n got: %#v", expected, args)
	}
}

func TestBuildRipgrepArgsContentModeAndContext(t *testing.T) {
	in := grepInput{Pattern: "foo", Context: 2}
	args := buildRipgrepArgs(in, "content", "/repo")
	expected := []string{"--sort", "path", "-n", "-C2", "foo", "/repo"}
	if !reflect.DeepEqual(args, expected) {
		t.Fatalf("unexpected args:\nwant: %#v\n got: %#v", expected, args)
	}
}

func TestLimitGrepOutputByLines(t *testing.T) {
	input := "a\nb\nc\nd"
	out := limitGrepOutput(input, 2, 100)
	wantPrefix := "a\nb\n... (truncated to 2 lines)"
	if !strings.HasPrefix(out, wantPrefix) {
		t.Fatalf("unexpected output prefix:\nwant: %q\n got: %q", wantPrefix, out)
	}
}

func TestLimitGrepOutputIncludesOverflowMetadata(t *testing.T) {
	input := "a\nb\nc\nd"
	out := limitGrepOutput(input, 2, 100)
	if !strings.Contains(out, "[truncation_metadata]") {
		t.Fatalf("expected truncation metadata in output, got: %q", out)
	}
	if !strings.Contains(out, "reason: exceeded_output_limits") {
		t.Fatalf("expected deterministic reason value, got: %q", out)
	}
	if !strings.Contains(out, "returned_bytes:") {
		t.Fatalf("expected returned_bytes metadata, got: %q", out)
	}
	if !strings.Contains(out, "artifact_available: true") {
		t.Fatalf("expected artifact availability metadata, got: %q", out)
	}
	if !strings.Contains(out, "artifact_path:") {
		t.Fatalf("expected artifact path in output, got: %q", out)
	}
}

func TestIsGrepOutputTruncated(t *testing.T) {
	if !isGrepOutputTruncated("a\nb\nc", 2, 100) {
		t.Fatalf("expected line-based truncation")
	}
	if !isGrepOutputTruncated("123456", 0, 5) {
		t.Fatalf("expected char-based truncation")
	}
	if isGrepOutputTruncated("a\nb", 5, 100) {
		t.Fatalf("did not expect truncation")
	}
}

func TestGrepNoMatchesIncludesStructuredBlock(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(file, []byte("hello\nworld\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	in, _ := json.Marshal(grepInput{Pattern: "nomatch", Path: dir})
	res, err := (&GrepTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}
	if !strings.Contains(res.Content, "[grep_result]") {
		t.Fatalf("expected structured no-match block, got: %s", res.Content)
	}
}
