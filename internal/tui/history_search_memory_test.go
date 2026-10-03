package tui

import "testing"

func TestHistorySearchRemembersLastSelectedPromptAcrossRelatedQueries(t *testing.T) {
	state := newHistorySearchState([]historySearchEntry{{text: "deploy release"}, {text: "deploy hotfix"}, {text: "open logs"}})
	state.setQuery("deploy")
	state.selected = 1
	state.rememberSelection()
	state.setQuery("deploy hot")
	entry, ok := state.selectedEntry()
	if !ok || entry.text != "deploy release" {
		t.Fatalf("expected related query to keep last selection, got %#v ok=%v", entry, ok)
	}
}

func TestHistorySearchStateInitializesSelectionMemory(t *testing.T) {
	state := newHistorySearchState([]historySearchEntry{{text: "alpha"}})
	if state.memory == nil {
		t.Fatalf("expected history selection memory map initialized")
	}
}
