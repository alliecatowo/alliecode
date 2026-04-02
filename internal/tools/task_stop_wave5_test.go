package tools

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskStopRejectsNonRunningTasks(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })
	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()
	taskAdapterUpdate("done-1", "completed", "ok", "")

	res, err := (&TaskStopTool{}).Execute(context.Background(), []byte(`{"task_id":"done-1"}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("task_stop execute error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected task_stop to reject non-running tasks")
	}
}
