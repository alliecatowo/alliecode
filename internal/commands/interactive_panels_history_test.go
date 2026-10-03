package commands

import "testing"

func TestInteractivePanelHistoryIncludesLatestAndEntryShow(t *testing.T) {
	panel := requirePanel(t, "history", &RuntimeState{HistoryEntries: []HistoryEntry{{ID: "h-1", Title: "Fix deploy", Summary: "Summary", Model: "gpt-4o-mini", CreatedAt: "2026-04-01T10:00:00Z"}}})
	if len(panel.HeaderIntents) == 0 {
		t.Fatalf("expected header intents for history panel")
	}
	want := map[string]bool{"/history latest": false, "/history show h-1": false}
	for _, item := range panel.Items {
		if _, ok := want[item.ApplyInput]; ok {
			want[item.ApplyInput] = true
		}
	}
	for apply, ok := range want {
		if !ok {
			t.Fatalf("expected history panel action %q", apply)
		}
	}
	for _, item := range panel.Items {
		if item.ApplyInput == "/history show h-1" && len(item.PreviewIntents) == 0 {
			t.Fatalf("expected structured history preview")
		}
	}
}
