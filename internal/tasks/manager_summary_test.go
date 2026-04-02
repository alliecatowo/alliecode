package tasks

import "testing"

func TestManagerStatusSummaryByOwner(t *testing.T) {
	mgr := NewManager(nil)
	mgr.tasks["a"] = &Task{ID: "a", Status: StatusRunning, Owner: "dev1"}
	mgr.tasks["b"] = &Task{ID: "b", Status: StatusCompleted, Owner: "dev2"}

	s := mgr.StatusSummaryByOwner("dev1")
	if s.Total != 1 || s.ByStatus[StatusRunning] != 1 {
		t.Fatalf("unexpected owner summary: %+v", s)
	}
}
