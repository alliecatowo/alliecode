package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestStage7ModalEscapeClearsSearchLayer(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	if !app.openInteractiveCommandPanel("permissions") {
		t.Fatalf("expected command panel open")
	}
	app.startQuickOpen()
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if app.state != stateIdle {
		t.Fatalf("expected search escape to return idle")
	}
	if got := app.activeModalSurface(); got != modalSurfaceNone {
		t.Fatalf("expected no active modal surface after search close, got %q", got)
	}
}
