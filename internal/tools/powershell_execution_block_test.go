package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestPowerShellToolAppendsExecutionBlock(t *testing.T) {
	in, _ := json.Marshal(powerShellInput{Command: "printf 'ok'"})
	res, err := (&PowerShellTool{}).Execute(context.Background(), in, types.ToolContext{WorkingDir: t.TempDir()})
	if err != nil || res.IsError {
		t.Fatalf("powershell failed: err=%v content=%q", err, res.Content)
	}
	if !strings.Contains(res.Content, "[shell_execution]") {
		t.Fatalf("expected shell execution block: %s", res.Content)
	}
}
