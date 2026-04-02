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

func TestFileEditIncludesStructuredResultBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(path, []byte("hello old"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat fixture: %v", err)
	}
	meta := renderReadMetadata(path, stat.ModTime(), stat.Size(), false, 0, 2000)
	in, _ := json.Marshal(fileEditInput{FilePath: path, OldString: "old", NewString: "new"})
	ctx := types.ToolContext{WorkingDir: dir, Messages: []types.Message{{Role: types.RoleUser, Content: []types.ContentBlock{{Type: types.ContentToolResult, Content: meta}}}}}
	res, execErr := (&FileEditTool{}).Execute(context.Background(), in, ctx)
	if execErr != nil {
		t.Fatalf("execute error: %v", execErr)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}
	if !strings.Contains(res.Content, "[file_edit_result]") {
		t.Fatalf("expected structured block, got: %s", res.Content)
	}
}
