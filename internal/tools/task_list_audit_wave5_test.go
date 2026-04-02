package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskListIncludesAuditAndLifecycleMap(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })
	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()
	taskAdapterUpdate("t1", "running", "", "")
	taskAdapterUpdate("t2", "completed", "done", "")

	res, err := (&TaskListTool{}).Execute(context.Background(), []byte(`{}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("task_list failed: err=%v content=%q", err, res.Content)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["audit"] == nil {
		t.Fatalf("expected audit payload: %v", out)
	}
	if out["lifecycle"] == nil {
		t.Fatalf("expected lifecycle payload: %v", out)
	}
}
