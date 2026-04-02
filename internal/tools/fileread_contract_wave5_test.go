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

func TestFileReadUsesOneBasedLineNumbers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "one-based.txt")
	if err := os.WriteFile(path, []byte("a\nb\nc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	in, _ := json.Marshal(fileReadInput{FilePath: path, Offset: 2, Limit: 1})
	res, err := (&FileReadTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	if err != nil || res.IsError {
		t.Fatalf("read failed: err=%v content=%q", err, res.Content)
	}
	if !strings.Contains(res.Content, "2: b") {
		t.Fatalf("expected one-based line format, got: %s", res.Content)
	}
}

func TestFileReadDirectoryListingContract(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "file.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	in, _ := json.Marshal(fileReadInput{FilePath: dir, Offset: 1, Limit: 10})
	res, err := (&FileReadTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	if err != nil || res.IsError {
		t.Fatalf("read dir failed: err=%v content=%q", err, res.Content)
	}
	if !strings.Contains(res.Content, "sub/") || !strings.Contains(res.Content, "file.txt") {
		t.Fatalf("expected directory entries with trailing slash for dirs, got: %s", res.Content)
	}
	if !strings.Contains(res.Content, "is_directory: true") {
		t.Fatalf("expected directory runtime field, got: %s", res.Content)
	}
}

func TestFileReadBlocksSensitiveDevicePaths(t *testing.T) {
	in, _ := json.Marshal(fileReadInput{FilePath: "/dev/zero"})
	res, err := (&FileReadTool{}).Execute(context.Background(), in, types.ToolContext{})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if !res.IsError || !strings.Contains(strings.ToLower(res.Content), "blocked") {
		t.Fatalf("expected blocked device path error, got: %q", res.Content)
	}
}
