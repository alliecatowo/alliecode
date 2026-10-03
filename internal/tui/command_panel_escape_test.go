package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestCommandPanelEscapeClosesWithoutApplying(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	if !app.openInteractiveCommandPanel("doctor") {
		t.Fatalf("expected doctor panel to open")
	}
	if !app.commandPanel.active {
		t.Fatalf("expected doctor panel active")
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if app.commandPanel.active {
		t.Fatalf("expected command panel closed after escape")
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected no timeline output after escape, got %d rows", len(app.timeline))
	}
	if got := app.input.Value(); got != "/doctor" {
		t.Fatalf("expected command text preserved after escape, got %q", got)
	}
}

func TestCommandPanelLeftArrowClosesWithoutApplying(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	if !app.openInteractiveCommandPanel("doctor") {
		t.Fatalf("expected doctor panel to open")
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyLeft}, "left")
	if app.commandPanel.active {
		t.Fatalf("expected command panel closed after left arrow")
	}
}
