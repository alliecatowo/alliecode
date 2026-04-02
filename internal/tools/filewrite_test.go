package tools

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestFileWriteIncludesStructuredResultBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "new.txt")
	in, _ := json.Marshal(fileWriteInput{FilePath: path, Content: "hello"})
	res, err := (&FileWriteTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	if err != nil {
		t.Fatalf("execute error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}
	if !strings.Contains(res.Content, "[file_write_result]") {
		t.Fatalf("expected structured block, got: %s", res.Content)
	}
}
