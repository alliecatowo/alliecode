package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInputModeMachineTracksActiveModeTransitions(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	if app.inputMode != inputModeChat {
		t.Fatalf("expected initial chat mode, got %s", app.inputMode)
	}

	app = typeTestText(t, app, "/")
	if app.inputMode != inputModeSlash {
		t.Fatalf("expected slash mode, got %s", app.inputMode)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if app.inputMode != inputModeChat {
		t.Fatalf("expected escape to return chat mode, got %s", app.inputMode)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	if app.inputMode != inputModeQuickOpen {
		t.Fatalf("expected quick-open mode, got %s", app.inputMode)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlR}, "ctrl+r")
	if app.inputMode != inputModeHistorySearch {
		t.Fatalf("expected history-search mode, got %s", app.inputMode)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, "ctrl+f")
	if app.inputMode != inputModeSearch {
		t.Fatalf("expected timeline-search mode, got %s", app.inputMode)
	}
}

func TestInputModeMachinePermissionTakesPriority(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.permissionQueue = []permissionPromptRequest{{toolName: "bash", description: "execute shell command"}}
	app.ensurePermissionPromptVisible()
	if app.inputMode != inputModePermission {
		t.Fatalf("expected permission mode, got %s", app.inputMode)
	}
}

func TestInputModeMachineTabAndShiftTabStayWithinActiveOverlayMode(t *testing.T) {
	app := readySizedApp(t, 160, 28)

	app = typeTestText(t, app, "/")
	if app.inputMode != inputModeSlash {
		t.Fatalf("expected slash mode, got %s", app.inputMode)
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyTab}, "tab")
	if app.inputMode != inputModeSlash {
		t.Fatalf("expected tab to keep slash mode active, got %s", app.inputMode)
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if app.inputMode != inputModeSlash {
		t.Fatalf("expected shift+tab to keep slash mode active, got %s", app.inputMode)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	if app.inputMode != inputModeQuickOpen {
		t.Fatalf("expected quick-open mode, got %s", app.inputMode)
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyTab}, "tab")
	if app.inputMode != inputModeQuickOpen {
		t.Fatalf("expected tab to keep quick-open mode, got %s", app.inputMode)
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if app.inputMode != inputModeQuickOpen {
		t.Fatalf("expected shift+tab to keep quick-open mode, got %s", app.inputMode)
	}
}

func TestInputModeMachineShiftTabSequenceStaysWithinActiveOverlayMode(t *testing.T) {
	app := readySizedApp(t, 160, 28)

	app = typeTestText(t, app, "/")
	if app.inputMode != inputModeSlash {
		t.Fatalf("expected slash mode, got %s", app.inputMode)
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'', '[', 'Z'}}, "shift-tab-csi-z")
	if app.inputMode != inputModeSlash {
		t.Fatalf("expected shift+tab sequence to keep slash mode active, got %s", app.inputMode)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	if app.inputMode != inputModeQuickOpen {
		t.Fatalf("expected quick-open mode, got %s", app.inputMode)
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'', '[', '1', ';', '2', 'Z'}}, "shift-tab-csi-1-2-z")
	if app.inputMode != inputModeQuickOpen {
		t.Fatalf("expected shift+tab sequence to keep quick-open mode, got %s", app.inputMode)
	}
}
