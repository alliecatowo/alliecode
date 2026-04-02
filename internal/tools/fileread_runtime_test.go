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

func TestFileReadIncludesRuntimeBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "read.txt")
	if err := os.WriteFile(path, []byte("a\nb\nc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	in, _ := json.Marshal(fileReadInput{FilePath: path, Offset: 0, Limit: 2})
	res, err := (&FileReadTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	if err != nil || res.IsError {
		t.Fatalf("read failed: err=%v content=%q", err, res.Content)
	}
	if !strings.Contains(res.Content, "[file_read_result]") || !strings.Contains(res.Content, "returned_lines: 2") {
		t.Fatalf("expected read runtime contract, got: %s", res.Content)
	}
}
