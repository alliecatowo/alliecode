package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskListIncludesContractWave6(t *testing.T) {
	res, err := (&TaskListTool{}).Execute(context.Background(), []byte(`{}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("task_list failed: err=%v content=%q", err, res.Content)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["contract"] == nil {
		t.Fatalf("expected contract metadata")
	}
}
