package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskListToolExecuteReturnsFormattedTodos(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })

	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()
	taskAdapterUpdate("task-b", "completed", "ok", "")
	taskAdapterUpdate("task-a", "running", "", "")

	res, err := (&TaskListTool{}).Execute(context.Background(), []byte(`{}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected non-error result: %s", res.Content)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json output: %v", err)
	}
	if out["total"] != float64(2) {
		t.Fatalf("expected total=2, got %v", out["total"])
	}
	tasks, _ := out["tasks"].([]any)
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	t0 := tasks[0].(map[string]any)
	if t0["task_id"] != "task-a" {
		t.Fatalf("expected deterministic sort by task_id, got %v", t0["task_id"])
	}

	runRes, err := (&TaskListTool{}).Execute(context.Background(), []byte(`{"status":"running"}`), types.ToolContext{})
	if err != nil || runRes.IsError {
		t.Fatalf("status-filter list failed: err=%v content=%q", err, runRes.Content)
	}
	var runOut map[string]any
	if err := json.Unmarshal([]byte(runRes.Content), &runOut); err != nil {
		t.Fatalf("invalid filtered json output: %v", err)
	}
	runTasks, _ := runOut["tasks"].([]any)
	if len(runTasks) != 1 {
		t.Fatalf("expected 1 running task, got %d", len(runTasks))
	}

	badRes, err := (&TaskListTool{}).Execute(context.Background(), []byte(`{"status":"paused"}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("bad status returned error: %v", err)
	}
	if !badRes.IsError {
		t.Fatalf("expected bad status to be rejected")
	}
}
