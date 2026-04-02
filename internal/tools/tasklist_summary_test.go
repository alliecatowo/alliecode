package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskListToolIncludesSummary(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })

	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()

	taskAdapterUpdate("task-a", "running", "", "")
	taskAdapterSetFields("task-a", func(rec *taskAdapterRecord) { rec.Owner = "dev1" })
	taskAdapterUpdate("task-b", "completed", "ok", "")

	res, err := (&TaskListTool{}).Execute(context.Background(), []byte(`{}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("task_list failed: err=%v content=%q", err, res.Content)
	}

	var out struct {
		Summary map[string]any `json:"summary"`
	}
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out.Summary == nil {
		t.Fatalf("expected summary in output")
	}
}
