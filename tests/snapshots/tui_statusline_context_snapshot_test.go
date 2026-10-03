package snapshots_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSnapshot_QuickOpenContextPivotToTimeline(t *testing.T) {
	app := readyApp(t, 180, 30)
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, "ctrl+f")

	plain := stripANSI(app.View())
	got := strings.Join([]string{
		trimRight(mustFindLineContaining(t, plain, "mode:timeline")),
		trimRight(mustFindLineContaining(t, plain, "hints: search active | search mode timeline selection=")),
	}, "\n")
	const want = "mode:timeline  matches:0  jump:ctrl+n/ctrl+p up/down tab shift+tab pgup/pgd...  esc:exit\nhints: search active | search mode timeline selection=- | input mode: search"
	if got != want {
		t.Fatalf("quick-open to timeline context snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSnapshot_PermissionsPanelContextFromSlashQuery(t *testing.T) {
	app := readyApp(t, 180, 30)
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/perm")}, "type:/perm")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	plain := stripANSI(app.View())
	got := strings.Join([]string{
		strings.TrimSpace(mustFindLineContaining(t, plain, "hint: usage: /permissions")),
		strings.TrimSpace(mustFindLineContaining(t, plain, "input mode: command-panel  |  command panel /permissions")),
		strings.TrimSpace(mustFindLineContaining(t, plain, "┃ /permissions")),
	}, "\n")
	const want = "hint: usage: /permissions [status|get|list|summary|rules [list|add <rule>|remove <rule>]|denials|retry-denials|set <plan|default|auto|bypass>|plan|default|auto|bypass]  -  Get...\ninput mode: command-panel  |  command panel /permissions -> Permissions summary  |  enter/right apply  tab cycle  shift+tab reverse  esc close\n┃ /permissions"
	if got != want {
		t.Fatalf("permissions panel context snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSnapshot_StatuslineRuntimeHiddenForIdleChat(t *testing.T) {
	app := readyApp(t, 180, 30)
	plain := stripANSI(app.View())
	if strings.Contains(plain, "runtime: idle") {
		t.Fatalf("expected runtime pane hidden for idle chat by default\n%s", plain)
	}
	if !strings.Contains(plain, "input mode: chat") {
		t.Fatalf("expected input mode hint line visible\n%s", plain)
	}
}
