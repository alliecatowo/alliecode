package history

import (
	"testing"
	"time"
)

func TestBuildIndexExtendedFilters(t *testing.T) {
	events := []Event{
		{Type: EventCustom, SessionID: "s1", Project: "/repo/a", Display: "deploy", RuntimeSurface: "cli", CommandSurface: "repl", Tags: []string{"prod"}, Timestamp: time.Unix(1, 0).UTC()},
		{Type: EventCustom, SessionID: "s2", Project: "/repo/a", Display: "test", RuntimeSurface: "ide", CommandSurface: "panel", Tags: []string{"ci"}, Timestamp: time.Unix(2, 0).UTC()},
	}
	out := BuildIndex(events, IndexQuery{RuntimeSurface: "cli", CommandSurface: "repl", TagContains: "pro", Limit: 10})
	if len(out) != 1 || out[0].SessionID != "s1" {
		t.Fatalf("unexpected index output: %+v", out)
	}
}

func TestSummarizeIndex(t *testing.T) {
	items := []IndexRecord{{Project: "/repo/a", Type: EventCustom}, {Project: "/repo/a", Type: EventMessage}}
	summary := SummarizeIndex(items)
	if summary.Total != 2 || summary.Projects["/repo/a"] != 2 || summary.DistinctSessions != 0 {
		t.Fatalf("unexpected index summary: %+v", summary)
	}
}
