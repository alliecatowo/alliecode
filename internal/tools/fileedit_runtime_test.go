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

func TestFileEditRuntimeFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "runtime.txt")
	if err := os.WriteFile(path, []byte("alpha beta"), 0o644); err != nil {
		t.Fatal(err)
	}
	stat, _ := os.Stat(path)
	meta := renderReadMetadata(path, stat.ModTime(), stat.Size(), false, 0, 2000)
	in, _ := json.Marshal(fileEditInput{FilePath: path, OldString: "beta", NewString: "gamma"})
	ctx := types.ToolContext{WorkingDir: dir, Messages: []types.Message{{Role: types.RoleUser, Content: []types.ContentBlock{{Type: types.ContentToolResult, Content: meta}}}}}
	res, err := (&FileEditTool{}).Execute(context.Background(), in, ctx)
	if err != nil || res.IsError {
		t.Fatalf("edit failed: err=%v content=%q", err, res.Content)
	}
	for _, needle := range []string{"old_length:", "new_length:", "duration_ms:", "workspace_dir:"} {
		if !strings.Contains(res.Content, needle) {
			t.Fatalf("expected %q in output: %s", needle, res.Content)
		}
	}
}
