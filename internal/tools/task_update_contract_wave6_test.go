package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskUpdateIncludesContractWave6(t *testing.T) {
	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()
	taskAdapterUpdate("wave6-upd", "running", "", "")

	res, err := (&TaskUpdateTool{}).Execute(context.Background(), []byte(`{"task_id":"wave6-upd","status":"completed"}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("task_update failed: err=%v content=%q", err, res.Content)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["contract"] == nil {
		t.Fatalf("expected contract metadata")
	}
}
