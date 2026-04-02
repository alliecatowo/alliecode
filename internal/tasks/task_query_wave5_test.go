package tasks

import "testing"

func TestApplyTaskQueryIncludesScannedAndFilteredCounts(t *testing.T) {
	includeTerminal := false
	items := []Task{
		{ID: "a", Status: StatusRunning, Owner: "dev"},
		{ID: "b", Status: StatusCompleted, Owner: "dev"},
	}
	_, summary := ApplyTaskQuery(items, TaskQuery{Owner: "dev", IncludeTerminal: &includeTerminal})
	if summary.Scanned != 2 {
		t.Fatalf("expected scanned=2, got %d", summary.Scanned)
	}
	if summary.FilteredOut != 1 {
		t.Fatalf("expected filtered_out=1, got %d", summary.FilteredOut)
	}
}
