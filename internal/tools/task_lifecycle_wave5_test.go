package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskCreateIncludesLifecycleAndRuntime(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })

	res, err := (&TaskCreateTool{}).Execute(context.Background(), []byte(`{"subject":"s","description":"d"}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("task_create failed: err=%v content=%q", err, res.Content)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["lifecycle"] == nil || out["runtime"] == nil {
		t.Fatalf("expected lifecycle+runtime fields, got: %v", out)
	}
}

func TestTaskOutputIncludesLifecycleAndRuntimeAudit(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })
	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()
	taskAdapterUpdate("task-life", "completed", "ok", "")

	res, err := (&TaskOutputTool{}).Execute(context.Background(), []byte(`{"task_id":"task-life","block":false}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("task_output failed: err=%v content=%q", err, res.Content)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["lifecycle"] == nil || out["runtime"] == nil {
		t.Fatalf("expected lifecycle+runtime fields, got: %v", out)
	}
}
