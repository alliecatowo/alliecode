package tools

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestFileWriteRuntimeFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "runtime.txt")
	in, _ := json.Marshal(fileWriteInput{FilePath: path, Content: "hello"})
	res, err := (&FileWriteTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: dir})
	if err != nil || res.IsError {
		t.Fatalf("write failed: err=%v content=%q", err, res.Content)
	}
	for _, needle := range []string{"existed_before:", "previous_size:", "duration_ms:", "workspace_dir:"} {
		if !strings.Contains(res.Content, needle) {
			t.Fatalf("expected %q in output: %s", needle, res.Content)
		}
	}
}
