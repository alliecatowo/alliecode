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
	if !strings.Contains(app.renderSearchHintPane(), "enter applies the selected action") {
		t.Fatalf("expected quick-open hint pane")
	}
}

func TestSearchFooterHintsIncludeActiveSelectionContext(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.startQuickOpen()
	line := stripANSIForTest(app.renderSearchFooterHints())
	if !strings.Contains(line, "active:") {
		t.Fatalf("expected active quick-open footer hint, got %q", line)
	}

	app.searchMode = searchModeHistory
	app.history = newHistorySearchState([]historySearchEntry{{text: "deploy release\nverify"}})
	line = stripANSIForTest(app.renderSearchFooterHints())
	if !strings.Contains(line, "enter/right apply") {
		t.Fatalf("expected history footer hint controls, got %q", line)
	}
	if !strings.Contains(line, "left/esc close") {
		t.Fatalf("expected history footer close semantics, got %q", line)
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
	if !strings.Contains(pane, "selected: ship build") {
		t.Fatalf("expected history details pane selected-row context, got %q", pane)
	}
}
