package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModalStackRoutesKeysThroughTopSurface(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	if !app.openInteractiveCommandPanel("permissions") {
		t.Fatalf("expected command panel open")
	}
	app.startQuickOpen()
	if got := app.activeModalSurface(); got != modalSurfaceSearch {
		t.Fatalf("expected search to own key routing, got %q", got)
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if app.state != stateIdle {
		t.Fatalf("expected esc to close search first")
	}
	if got := app.activeModalSurface(); got != modalSurfaceNone {
		t.Fatalf("expected no active modal after search close, got %q", got)
	}
}
