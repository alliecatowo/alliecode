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

func TestFileEditCanCreateWhenOldStringEmptyAndMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.txt")
	in, _ := json.Marshal(fileEditInput{FilePath: path, OldString: "", NewString: "hello"})
	res, err := (&FileEditTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	if err != nil || res.IsError {
		t.Fatalf("edit create failed: err=%v content=%q", err, res.Content)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "hello" {
		t.Fatalf("unexpected file content: %q", string(data))
	}
	if !strings.Contains(res.Content, "created: true") {
		t.Fatalf("expected created runtime field: %s", res.Content)
	}
}

func TestFileEditRejectsNotebookFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "n.ipynb")
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	in, _ := json.Marshal(fileEditInput{FilePath: path, OldString: "{}", NewString: "[]"})
	res, err := (&FileEditTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	if err != nil {
		t.Fatalf("execute failed: %v", err)
	}
	if !res.IsError || !strings.Contains(strings.ToLower(res.Content), "notebook") {
		t.Fatalf("expected notebook rejection, got: %q", res.Content)
	}
}
