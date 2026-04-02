package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskGetAndUpdateReturnSummary(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })
	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()
	taskAdapterUpdate("x", "running", "", "")

	getRes, _ := (&TaskGetTool{}).Execute(context.Background(), []byte(`{"task_id":"x"}`), types.ToolContext{})
	var got map[string]any
	_ = json.Unmarshal([]byte(getRes.Content), &got)
	if got["summary"] == nil {
		t.Fatalf("expected summary in task_get response: %s", getRes.Content)
	}

	updRes, _ := (&TaskUpdateTool{}).Execute(context.Background(), []byte(`{"task_id":"x","status":"completed"}`), types.ToolContext{})
	_ = json.Unmarshal([]byte(updRes.Content), &got)
	if got["summary"] == nil {
		t.Fatalf("expected summary in task_update response: %s", updRes.Content)
	}
}

func TestTaskOutputIncludesRuntime(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })
	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{}
	taskAdapterMu.Unlock()
	taskAdapterUpdate("x", "completed", "ok", "")

	res, _ := (&TaskOutputTool{}).Execute(context.Background(), []byte(`{"task_id":"x"}`), types.ToolContext{})
	var got map[string]any
	_ = json.Unmarshal([]byte(res.Content), &got)
	if got["runtime"] == nil {
		t.Fatalf("expected runtime in task_output response: %s", res.Content)
	}
}
