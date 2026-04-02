package tasks

import "testing"

func TestApplyTaskQuery(t *testing.T) {
	items := []Task{{ID: "a", Status: StatusRunning, Owner: "dev1"}, {ID: "b", Status: StatusCompleted, Owner: "dev1"}, {ID: "c", Status: StatusRunning, Owner: "dev2"}}
	includeTerminal := false
	filtered, summary := ApplyTaskQuery(items, TaskQuery{Owner: "dev1", IncludeTerminal: &includeTerminal})
	if len(filtered) != 1 || filtered[0].ID != "a" {
		t.Fatalf("unexpected filtered tasks: %+v", filtered)
	}
	if summary.Matched != 1 || summary.ByStatus[StatusRunning] != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}
