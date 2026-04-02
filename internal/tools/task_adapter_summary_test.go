package tools

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/tasks"
)

func TestTasksStatusSummaryFromList(t *testing.T) {
	s := tasksStatusSummaryFromList([]tasks.Task{{ID: "a", Status: tasks.StatusRunning}, {ID: "b", Status: tasks.StatusCompleted}})
	if s.Total != 2 || s.ByStatus[tasks.StatusCompleted] != 1 {
		t.Fatalf("unexpected summary: %+v", s)
	}
}

func TestTaskAdapterSummaryUsesAdapterState(t *testing.T) {
	oldMgr := agentTaskManager
	agentTaskManager = nil
	t.Cleanup(func() { agentTaskManager = oldMgr })

	taskAdapterMu.Lock()
	taskAdapterRecords = map[string]taskAdapterRecord{"a": {Status: "running"}}
	taskAdapterMu.Unlock()

	s := taskAdapterSummary(context.Background())
	if s.Total != 1 {
		t.Fatalf("expected one record, got %+v", s)
	}
}
