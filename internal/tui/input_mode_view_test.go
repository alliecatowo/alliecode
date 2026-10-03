package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInputModeIndicatorShowsModeContextAndHint(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	view := stripANSIForTest(app.renderInputModeIndicator())
	if !strings.Contains(view, "mode: quick-open") {
		t.Fatalf("expected quick-open mode indicator, got %q", view)
	}
	if !strings.Contains(view, "enter/right apply") {
		t.Fatalf("expected compact mode hint, got %q", view)
	}
	if !strings.Contains(view, "shift+tab reverse") {
		t.Fatalf("expected reverse-cycle hint in mode indicator, got %q", view)
	}
	if !strings.Contains(view, "esc close") {
		t.Fatalf("expected close semantics in mode indicator, got %q", view)
	}
}
