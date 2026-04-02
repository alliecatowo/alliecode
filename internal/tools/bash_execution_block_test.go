package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestBashToolAppendsExecutionBlock(t *testing.T) {
	in, _ := json.Marshal(bashInput{Command: "printf 'ok'"})
	res, err := (&BashTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: t.TempDir()})
	if err != nil || res.IsError {
		t.Fatalf("bash failed: err=%v content=%q", err, res.Content)
	}
	if !strings.Contains(res.Content, "[shell_execution]") {
		t.Fatalf("expected shell execution block: %s", res.Content)
	}
}
