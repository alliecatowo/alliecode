package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func TestRenderStatusHintsForInteractiveModes(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.state = stateSearch
	if !strings.Contains(app.renderStatusHints(), "search active") {
		t.Fatalf("expected search hint marker")
	}
	app.state = stateIdle
	app.slashAutocomplete.items = []commands.Suggestion{{Name: "model"}}
	if !strings.Contains(app.renderStatusHints(), "command palette") {
		t.Fatalf("expected command palette hint marker")
	}
}

func TestRenderStatusRuntimePanesIncludesRuntimeAndContext(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.searchTimelineQuery = "errors"
	app.state = stateSearch
	view := app.renderStatusRuntimePanes()
	if !strings.Contains(view, "runtime: state=search") {
		t.Fatalf("expected runtime pane to include state, got %q", view)
	}
	if !strings.Contains(view, "context:") {
		t.Fatalf("expected context pane, got %q", view)
	}
	if !strings.Contains(view, "searches:") {
		t.Fatalf("expected search memory pane, got %q", view)
	}
}
