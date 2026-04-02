package tui

import (
	"strings"
	"testing"
)

func TestSearchHintPaneChangesByMode(t *testing.T) {
	app := New(Config{})
	app.searchMode = searchModeTimeline
	if !strings.Contains(app.renderSearchHintPane(), "type to filter") {
		t.Fatalf("expected timeline hint pane")
	}
	app.searchMode = searchModeQuickOpen
	if !strings.Contains(app.renderSearchHintPane(), "enter to run") {
		t.Fatalf("expected quick-open hint pane")
	}
}

func TestSearchDetailsPaneShowsHistoryPreview(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.searchMode = searchModeHistory
	app.history = newHistorySearchState([]historySearchEntry{{text: "ship build\nverify rollout"}})
	pane := app.renderSearchDetailsPane()
	if !strings.Contains(pane, "history details:") {
		t.Fatalf("expected history details pane, got %q", pane)
	}
}
