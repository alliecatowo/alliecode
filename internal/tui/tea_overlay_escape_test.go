package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTeaOverlayEscapeModelPickerReturnsChatAndPreservesStagedCommand(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app = typeTestText(t, app, "/model ")
	if !app.modelPickerActive() {
		t.Fatalf("expected model picker active")
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if app.modelPickerActive() {
		t.Fatalf("expected model picker closed after escape")
	}
	if app.inputMode != inputModeChat {
		t.Fatalf("expected chat mode after model picker escape, got %s", app.inputMode)
	}
	if got := app.input.Value(); got != "/model " {
		t.Fatalf("expected staged /model input preserved, got %q", got)
	}
}

func TestTeaOverlayEscapeSearchModesReturnToChat(t *testing.T) {
	app := readySizedApp(t, 180, 30)

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if app.inputMode != inputModeChat || app.state != stateIdle {
		t.Fatalf("expected quick-open escape to return idle chat, state=%d mode=%s", app.state, app.inputMode)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlR}, "ctrl+r")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if app.inputMode != inputModeChat || app.state != stateIdle {
		t.Fatalf("expected history escape to return idle chat, state=%d mode=%s", app.state, app.inputMode)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, "ctrl+f")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if app.inputMode != inputModeChat || app.state != stateIdle {
		t.Fatalf("expected timeline escape to return idle chat, state=%d mode=%s", app.state, app.inputMode)
	}
}


func TestTeaOverlayEscapeSlashAndReferenceReturnToChat(t *testing.T) {
	app := readySizedApp(t, 180, 30)

	app = typeTestText(t, app, "/status")
	if app.inputMode != inputModeSlash {
		t.Fatalf("expected slash mode active, got %s", app.inputMode)
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if app.inputMode != inputModeChat {
		t.Fatalf("expected slash escape to restore chat mode, got %s", app.inputMode)
	}
	if got := app.input.Value(); got != "/status" {
		t.Fatalf("expected slash input preserved, got %q", got)
	}

	app = typeTestText(t, app, " @READ")
	if app.inputMode != inputModeReference {
		t.Fatalf("expected reference mode active, got %s", app.inputMode)
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if app.inputMode != inputModeChat {
		t.Fatalf("expected reference escape to restore chat mode, got %s", app.inputMode)
	}
}
