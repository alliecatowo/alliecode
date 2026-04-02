package tools

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestFileWriteIncludesWriteTypeField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.txt")
	in, _ := json.Marshal(fileWriteInput{FilePath: path, Content: "x"})
	res, err := (&FileWriteTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	if err != nil || res.IsError {
		t.Fatalf("write failed: err=%v content=%q", err, res.Content)
	}
	if !strings.Contains(res.Content, "write_type: create") {
		t.Fatalf("expected write_type=create, got %s", res.Content)
	}
}
