package snapshots_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSnapshot_Stage5QuickOpenContextAndModeHints(t *testing.T) {
	app := readyApp(t, 180, 30)
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")

	plain := stripANSI(app.View())
	got := strings.Join([]string{
		trimRight(mustFindLineContaining(t, plain, "hints: search active | search mode quick-open selection=")),
		trimRight(mustFindLineContaining(t, plain, "input mode: quick-open  |  search mode quick-open selection=")),
	}, "\n")
	const want = "hints: search active | search mode quick-open selection=Search timeline | input mode: quick-open\ninput mode: quick-open  |  search mode quick-open selection=Search timeline  |  enter/right apply  tab cycle  shift+tab reverse  esc close"
	if got != want {
		t.Fatalf("stage5 quick-open context snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSnapshot_Stage5CommandPanelContextAndModeHints(t *testing.T) {
	app := readyApp(t, 180, 30)
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/perm")}, "type:/perm")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	plain := stripANSI(app.View())
	got := strings.Join([]string{
		strings.TrimSpace(mustFindLineContaining(t, plain, "hint: usage: /permissions")),
		strings.TrimSpace(mustFindLineContaining(t, plain, "input mode: command-panel  |  command panel /permissions")),
	}, "\n")
	const want = "hint: usage: /permissions [status|get|list|summary|rules [list|add <rule>|remove <rule>]|denials|retry-denials|set <plan|default|auto|bypass>|plan|default|auto|bypass]  -  Get...\ninput mode: command-panel  |  command panel /permissions -> Permissions summary  |  enter/right apply  tab cycle  shift+tab reverse  esc close"
	if got != want {
		t.Fatalf("stage5 command panel context snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
