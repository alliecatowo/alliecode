package tui

import "testing"

func TestHistorySearchSelectedSummaryFallback(t *testing.T) {
	state := newHistorySearchState([]historySearchEntry{{text: "alpha"}})
	state.selected = -1
	if got := state.selectedSummary(20); got == "" {
		t.Fatalf("expected non-empty fallback summary")
	}
}

func TestHistorySearchSelectedDetailLinesIncludesPreview(t *testing.T) {
	state := newHistorySearchState([]historySearchEntry{{text: "line one\nline two\nline three\nline four"}})
	lines := state.selectedDetailLines(80)
	if len(lines) < 3 {
		t.Fatalf("expected details pane lines, got %#v", lines)
	}
	if lines[0] != "history details:" {
		t.Fatalf("expected details heading, got %q", lines[0])
	}
}

func TestHistorySearchSelectionMemoryByQuery(t *testing.T) {
	state := newHistorySearchState([]historySearchEntry{{text: "deploy api"}, {text: "deploy web"}})
	state.setQuery("web")
	state.moveSelection(1)
	state.setQuery("web")
	if state.selected < 0 {
		t.Fatalf("expected remembered selection in repeated query")
	}
}
