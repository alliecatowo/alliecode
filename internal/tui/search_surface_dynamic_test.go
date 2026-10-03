package tui

import (
	"strings"
	"testing"
)

func TestSearchSurfaceDoesNotPadToLegacyStableHeight(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.searchMode = searchModeTimeline
	app.searchQuery = ""
	view := stripANSIForTest(app.renderSearchLine())
	if lines := visualLineCount(view, 118); lines >= 14 {
		t.Fatalf("expected dynamic search height below legacy stable block, got %d lines", lines)
	}
}

func TestSearchSurfaceQuickOpenRendersDetailsBeforeListRows(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.startQuickOpen()
	view := stripANSIForTest(app.renderSearchLine())
	detail := strings.Index(view, "quick-open preview:")
	list := strings.Index(view, "Navigation:")
	if detail < 0 || list < 0 {
		t.Fatalf("expected quick-open detail and list sections, got %q", view)
	}
	if detail > list {
		t.Fatalf("expected detail pane before list rows, got %q", view)
	}
}
