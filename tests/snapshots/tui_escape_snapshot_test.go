package snapshots_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSnapshot_PermissionsPanelEscapeRestoresInputAndClosesPanel(t *testing.T) {
	app := readyApp(t, 180, 30)
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/perm")}, "type:/perm")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")

	plain := stripANSI(app.View())
	got := strings.Join([]string{
		trimRight(mustFindLineContaining(t, plain, "input mode: chat  |  input enter send  shift+tab history  ctrl+j newline")),
		trimRight(mustFindLineContaining(t, plain, "┃ /permissions")),
	}, "\n")
	const want = "input mode: chat  |  input enter send  shift+tab history  ctrl+j newline\n┃ /permissions"
	if got != want {
		t.Fatalf("permissions panel escape snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
	if strings.Contains(plain, "permissions panel: /permissions") {
		t.Fatalf("expected panel to be closed, got:\n%s", plain)
	}
}

func TestSnapshot_ModelPickerEscapeLeavesStagedCommand(t *testing.T) {
	app := readyApp(t, 180, 30)
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/model ")}, "type:/model ")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")

	plain := stripANSI(app.View())
	got := strings.Join([]string{
		trimRight(mustFindLineContaining(t, plain, "input mode: chat  |  input enter send  shift+tab history  ctrl+j newline")),
		trimRight(mustFindLineContaining(t, plain, "hint: /model [provider/model|model|list [provider|all]|doctor|repair [provider/model]]")),
	}, "\n")
	const want = "input mode: chat  |  input enter send  shift+tab history  ctrl+j newline\nhint: /model [provider/model|model|list [provider|all]|doctor|repair [provider/model]]  -  Get or set active model for this session  (add args, then enter)"
	if got != want {
		t.Fatalf("model picker escape snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
	if strings.Contains(plain, "model picker: /model") {
		t.Fatalf("expected model picker to be closed, got:\n%s", plain)
	}
}
