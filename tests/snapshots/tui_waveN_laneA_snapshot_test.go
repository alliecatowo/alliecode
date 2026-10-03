package snapshots_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSnapshot_WaveNLaneA_IdlePrimaryStatusAndCompactBuddy(t *testing.T) {
	app := readyApp(t, 120, 26)
	plain := stripANSI(app.View())
	if !strings.Contains(plain, "input chat") {
		t.Fatalf("expected primary status to include chat input mode, got:\n%s", plain)
	}
	if strings.Contains(plain, "hints:") || strings.Contains(plain, "runtime: ") {
		t.Fatalf("expected idle view without contextual status clutter, got:\n%s", plain)
	}
	if !strings.Contains(plain, "<°~°> buddy") {
		t.Fatalf("expected compact buddy line in idle view, got:\n%s", plain)
	}
}

func TestSnapshot_WaveNLaneA_SlashDrawerDocksNearComposer(t *testing.T) {
	app := readyApp(t, 140, 30)
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}, "type:/")
	plain := stripANSI(app.View())
	drawerIdx := mustFindLineIndexContaining(t, plain, "commands: /")
	composerIdx := mustFindLineIndexContaining(t, plain, "┃ /")
	buddyIdx := mustFindLineIndexContaining(t, plain, "`-vvvv-`")
	if !(drawerIdx < composerIdx && composerIdx < buddyIdx) {
		t.Fatalf("expected slash drawer docked above composer and below transcript, got:\n%s", plain)
	}
}

func TestSnapshot_WaveNLaneA_PanelCloseRestoresInputState(t *testing.T) {
	app := readyApp(t, 180, 30)
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/perm")}, "type:/perm")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	plain := stripANSI(app.View())
	if !strings.Contains(plain, "┃ /permissions") {
		t.Fatalf("expected previous slash input restored after panel close, got:\n%s", plain)
	}
	if strings.Contains(plain, "drawer: permissions panel") {
		t.Fatalf("expected command panel closed after escape, got:\n%s", plain)
	}
}

func TestSnapshot_WaveNLaneA_SearchKeepsContextualStatusVisible(t *testing.T) {
	app := readyApp(t, 160, 28)
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	plain := stripANSI(app.View())
	if !strings.Contains(plain, "mode:quick-open") {
		t.Fatalf("expected quick-open search mode line, got:\n%s", plain)
	}
	if !strings.Contains(plain, "hints: search active") {
		t.Fatalf("expected contextual hints while search active, got:\n%s", plain)
	}
	if !strings.Contains(plain, "runtime: search") {
		t.Fatalf("expected runtime pane while search active, got:\n%s", plain)
	}
}
