package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestGlobIncludesStructuredResultBlock(t *testing.T) {
	dir := t.TempDir()
	in, _ := json.Marshal(globInput{Pattern: "*.missing", Path: dir})
	res, err := (&GlobTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}
	if !strings.Contains(res.Content, "[glob_result]") {
		t.Fatalf("expected structured block, got: %s", res.Content)
	}

	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	in, _ = json.Marshal(globInput{Pattern: "*.txt", Path: dir})
	res, err = (&GlobTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if !strings.Contains(res.Content, "count: 1") {
		t.Fatalf("expected count metadata, got: %s", res.Content)
	}
}

func TestGlobSortAndFormatDeterministicTieBreakByPath(t *testing.T) {
	dir := t.TempDir()
	tieTime := time.Now().Add(-1 * time.Hour)
	paths := []string{
		filepath.Join(dir, "b.txt"),
		filepath.Join(dir, "a.txt"),
	}
	for _, p := range paths {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
		if err := os.Chtimes(p, tieTime, tieTime); err != nil {
			t.Fatalf("chtimes fixture: %v", err)
		}
	}
	sort.Strings(paths)
	res, err := (&GlobTool{}).sortAndFormat([]string{paths[1], paths[0]})
	if err != nil {
		t.Fatalf("sortAndFormat error: %v", err)
	}
	if !strings.Contains(res.Content, paths[0]+"\n"+paths[1]) {
		t.Fatalf("expected path tie-break ordering, got: %s", res.Content)
	}
}

func TestLimitGlobOutputIncludesTruncationMetadata(t *testing.T) {
	input := "a\nb\nc"
	out := limitGlobOutput(input, 2, 100)
	if !strings.Contains(out, "[truncation_metadata]") {
		t.Fatalf("expected truncation metadata, got: %q", out)
	}
	if !strings.Contains(out, "reason: exceeded_output_limits") {
		t.Fatalf("expected deterministic reason value, got: %q", out)
	}
	if !strings.Contains(out, "returned_bytes:") {
		t.Fatalf("expected returned_bytes metadata, got: %q", out)
	}
}

func TestLimitGlobOutputNoTruncationReturnsUnchanged(t *testing.T) {
	input := "one\ntwo"
	out := limitGlobOutput(input, 10, 100)
	if out != input {
		t.Fatalf("expected unchanged output, got: %q", out)
	}
}

func TestGlobResultCountRemainsTotalWhenOutputTruncated(t *testing.T) {
	dir := t.TempDir()
	var many []string
	for i := 0; i < 600; i++ {
		p := filepath.Join(dir, "f"+strconv.Itoa(i)+".txt")
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
		many = append(many, p)
	}
	res, err := (&GlobTool{}).sortAndFormat(many)
	if err != nil {
		t.Fatalf("sortAndFormat error: %v", err)
	}
	if !strings.Contains(res.Content, "[truncation_metadata]") {
		t.Fatalf("expected truncation metadata, got: %s", res.Content)
	}
	if !strings.Contains(res.Content, "count: 600") {
		t.Fatalf("expected total count metadata, got: %s", res.Content)
	}
}
