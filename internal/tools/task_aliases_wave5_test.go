package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskGetAcceptsTaskIdAlias(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })
	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()
	taskAdapterUpdate("alias-1", "running", "", "")

	res, err := (&TaskGetTool{}).Execute(context.Background(), []byte(`{"taskId":"alias-1"}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("task_get failed: err=%v content=%q", err, res.Content)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil || out["task"] == nil {
		t.Fatalf("expected task in response, err=%v out=%v", err, out)
	}
}

func TestTaskStopAcceptsShellIdAlias(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })
	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()
	taskAdapterUpdate("alias-stop", "running", "", "")

	res, err := (&TaskStopTool{}).Execute(context.Background(), []byte(`{"shell_id":"alias-stop"}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("task_stop failed: err=%v content=%q", err, res.Content)
	}
}
