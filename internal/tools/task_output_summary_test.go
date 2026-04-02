package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskOutputIncludesSummaryForResolvedTask(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })

	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()
	taskAdapterUpdate("task-7", "completed", "done", "")

	res, err := (&TaskOutputTool{}).Execute(context.Background(), []byte(`{"task_id":"task-7"}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("task_output failed: err=%v content=%q", err, res.Content)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["summary"] == nil {
		t.Fatalf("expected summary field")
	}
}
