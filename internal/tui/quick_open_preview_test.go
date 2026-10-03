package tui

import (
	"strings"
	"testing"
)

func TestQuickOpenPreviewIncludesSelectionMetadata(t *testing.T) {
	state := newQuickOpenState([]quickOpenItem{{label: "History search", detail: "Open prompt history", value: "search.history", status: "history"}})
	lines := state.selectedDetailLines(80)
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "selection: 1/1") {
		t.Fatalf("expected preview selection metadata, got %q", joined)
	}
}

func TestQuickOpenStateInitializesSelectionMemory(t *testing.T) {
	state := newQuickOpenState([]quickOpenItem{{label: "History search", value: "search.history"}})
	if state.memory == nil {
		t.Fatalf("expected quick-open selection memory map initialized")
	}
}
