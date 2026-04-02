package tasks

import "testing"

func TestManagerQueryAndSummary(t *testing.T) {
	mgr := NewManager(nil)
	mgr.tasks["a"] = &Task{ID: "a", Status: StatusRunning, Owner: "dev1"}
	mgr.tasks["b"] = &Task{ID: "b", Status: StatusCompleted, Owner: "dev1"}
	includeTerminal := false
	q := TaskQuery{Owner: "dev1", IncludeTerminal: &includeTerminal}
	items := mgr.Query(q)
	if len(items) != 1 || items[0].ID != "a" {
		t.Fatalf("unexpected query items: %+v", items)
	}
	s := mgr.QuerySummary(q)
	if s.Matched != 1 || s.ByStatus[StatusRunning] != 1 {
		t.Fatalf("unexpected query summary: %+v", s)
	}
}
