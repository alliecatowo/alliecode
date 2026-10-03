package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestResolveKeyActionUsesModeStack(t *testing.T) {
	app := New(Config{})
	app.syncInputMode()

	if got := app.resolveKeyAction(tea.KeyMsg{Type: tea.KeyCtrlF}); got != "search:timeline" {
		t.Fatalf("expected ctrl+f to resolve timeline search, got %q", got)
	}

	app.setState(stateSearch)
	app.setSearchModeValue(searchModeQuickOpen)
	app.syncInputMode()
	if got := app.resolveKeyAction(tea.KeyMsg{Type: tea.KeyCtrlR}); got != "search:history" {
		t.Fatalf("expected ctrl+r in quick-open mode to resolve history search, got %q", got)
	}
}

func TestHandleGlobalSearchAction(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.height = 40
	app.ready = true
	app.syncInputMode()

	if !app.handleGlobalSearchAction(tea.KeyMsg{Type: tea.KeyCtrlO}) {
		t.Fatalf("expected ctrl+o to be handled")
	}
	if app.stateValue() != stateSearch || app.searchModeValue() != searchModeQuickOpen {
		t.Fatalf("expected quick-open state transition, got state=%d mode=%d", app.stateValue(), app.searchModeValue())
	}

	if !app.handleGlobalSearchAction(tea.KeyMsg{Type: tea.KeyCtrlR}) {
		t.Fatalf("expected ctrl+r to be handled")
	}
	if app.searchModeValue() != searchModeHistory {
		t.Fatalf("expected history search mode, got %d", app.searchModeValue())
	}
}
