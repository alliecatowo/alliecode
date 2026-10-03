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
	app.searchMode = searchModeTimeline
	app.syncInputMode()
	hints := stripANSIForTest(app.renderStatusHints())
	if !strings.Contains(hints, "search active") {
		t.Fatalf("expected search hint marker")
	}
	app.state = stateIdle
	app.slashAutocomplete.items = []commands.Suggestion{{Name: "model"}}
	app.syncInputMode()
	if !strings.Contains(stripANSIForTest(app.renderStatusHints()), "command /model") {
		t.Fatalf("expected command palette hint marker")
	}
}

func TestRenderStatusRuntimePanesIncludesRuntimeAndContext(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.searchTimelineQuery = "errors"
	app.state = stateSearch
	app.searchMode = searchModeTimeline
	app.syncInputMode()
	view := stripANSIForTest(app.renderStatusRuntimePanes())
	if !strings.Contains(view, "runtime: search") {
		t.Fatalf("expected runtime pane to include state, got %q", view)
	}
	if !strings.Contains(view, "context:") {
		t.Fatalf("expected context pane, got %q", view)
	}
	if !strings.Contains(view, "searches:") {
		t.Fatalf("expected search memory pane, got %q", view)
	}
}

func TestRenderStatusHintsShowsLoginGuidanceWhenProviderNotReady(t *testing.T) {
	app := New(Config{InitialState: commands.RuntimeState{ProviderName: "anthropic", Model: "claude-opus-4-20250514", LoggedIn: true, AuthProvider: "openai"}})
	app.width = 120
	app.syncInputMode()

	if !strings.Contains(app.renderStatusHints(), "/login provider anthropic") {
		t.Fatalf("expected login guidance in status hints, got %q", app.renderStatusHints())
	}
}
