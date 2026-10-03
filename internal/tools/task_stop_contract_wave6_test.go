package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskStopIncludesContractWave6(t *testing.T) {
	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()
	taskAdapterUpdate("wave6-stop", "running", "", "")
	res, err := (&TaskStopTool{}).Execute(context.Background(), []byte(`{"task_id":"wave6-stop"}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("task_stop failed: err=%v content=%q", err, res.Content)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["contract"] == nil {
		t.Fatalf("expected contract metadata")
	}
}
