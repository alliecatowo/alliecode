package snapshots_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSnapshot_WaveMLane3_MixedPanelCycleReturnsToChat(t *testing.T) {
	app := readyApp(t, 120, 24)
	for i := 0; i < 6; i++ {
		app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")}, "type:hello")
		app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	}

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, "ctrl+f")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlR}, "ctrl+r")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")

	plain := stripANSI(app.View())
	for _, needle := range []string{"input mode: chat", "Type a message"} {
		if !strings.Contains(plain, needle) {
			t.Fatalf("expected %q after mixed panel cycle", needle)
		}
	}
	for _, needle := range []string{"mode:quick-open", "mode:history-search"} {
		if strings.Contains(plain, needle) {
			t.Fatalf("expected %q to be absent after returning to chat", needle)
		}
	}
}

func TestSnapshot_WaveMLane3_SlashToModelPickerEscLeavesStagedInput(t *testing.T) {
	app := readyApp(t, 180, 26)
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/model")}, "type:/model")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")

	plain := stripANSI(app.View())
	if !strings.Contains(plain, "┃ /model ") {
		t.Fatalf("expected staged /model input after picker escape")
	}
	if strings.Contains(plain, "model picker: /model") {
		t.Fatalf("expected model picker panel closed after escape")
	}
}

func TestSnapshot_WaveMLane3_TimelineSearchChromeDuringLongerTranscript(t *testing.T) {
	app := readyApp(t, 120, 24)
	for i := 0; i < 12; i++ {
		app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("line")}, "type:line")
		app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	}
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, "ctrl+f")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("line")}, "type:line")

	plain := stripANSI(app.View())
	for _, needle := range []string{"mode:timeline", "matches:", "jump:"} {
		if !strings.Contains(plain, needle) {
			t.Fatalf("expected %q in timeline search chrome", needle)
		}
	}
}
