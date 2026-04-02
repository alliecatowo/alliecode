package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestLSToolExecuteListsEntries(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}

	input, _ := json.Marshal(lsInput{})
	res, err := (&LSTool{}).Execute(context.Background(), input, types.ToolContext{WorkingDir: dir})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected non-error result: %s", res.Content)
	}
	if !strings.Contains(res.Content, "a.txt") {
		t.Fatalf("expected file listing, got: %q", res.Content)
	}
	if !strings.Contains(res.Content, "sub/") {
		t.Fatalf("expected directory listing with slash, got: %q", res.Content)
	}
}

func TestLSToolExecuteEmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	input, _ := json.Marshal(lsInput{})
	res, err := (&LSTool{}).Execute(context.Background(), input, types.ToolContext{WorkingDir: dir})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if res.Content != "(empty directory)" {
		t.Fatalf("unexpected content: %q", res.Content)
	}
}
