package e2e_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestE2EShiftedVariantsQuickOpenReverseAliases(t *testing.T) {
	app := e2eReadyApp(t, 180, 30)
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO})
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyTab})
	selectedAfterTab := e2eFindLineContaining(e2ePlainView(app), "selected:")

	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', 'Z'}})
	viewZ := e2ePlainView(app)
	if !strings.Contains(viewZ, "input mode: quick-open") {
		t.Fatalf("expected quick-open mode after CSI Z shift-tab alias")
	}
	selectedAfterCSIz := e2eFindLineContaining(viewZ, "selected:")
	if selectedAfterCSIz == selectedAfterTab {
		t.Fatalf("expected CSI Z shift-tab alias to reverse selection")
	}

	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', '1', ';', '2', 'Z'}})
	viewCSI12Z := e2ePlainView(app)
	if !strings.Contains(viewCSI12Z, "input mode: quick-open") {
		t.Fatalf("expected quick-open mode after CSI 1;2Z shift-tab alias")
	}
}

func TestE2EShiftedVariantsSlashAndModelPickerReverseAliases(t *testing.T) {
	app := e2eReadyApp(t, 180, 30)
	app = e2eTypeText(t, app, "/")
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyTab})
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', '9', ';', '2', 'u'}})
	slashView := e2ePlainView(app)
	if !strings.Contains(slashView, "input mode: slash") || !strings.Contains(slashView, "commands: /") {
		t.Fatalf("expected slash drawer to stay open on CSI 9;2u shift-tab alias, got:\n%s", slashView)
	}

	app = e2eReadyApp(t, 180, 30)
	app = e2eTypeText(t, app, "/model ")
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyTab})
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', '1', ';', '2', 'Z'}})
	pickerView := e2ePlainView(app)
	if !strings.Contains(pickerView, "model picker: /model") || !strings.Contains(pickerView, "input mode: model-picker") {
		t.Fatalf("expected model picker to stay open on CSI 1;2Z shift-tab alias, got:\n%s", pickerView)
	}
}

func TestE2EShiftedVariantsShiftEnterAliasInsertsNewlineWithoutSubmit(t *testing.T) {
	app := e2eReadyApp(t, 180, 30)
	app = e2eTypeText(t, app, "hello")
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', '1', '3', ';', '2', 'u'}})
	view := e2ePlainView(app)
	if strings.Contains(view, "runtime selection is incomplete") {
		t.Fatalf("expected shift-enter alias to avoid submit path, got:\n%s", view)
	}
	if !strings.Contains(view, "hello") {
		t.Fatalf("expected input to retain text after shift-enter alias, got:\n%s", view)
	}
}

func e2eFindLineContaining(view, needle string) string {
	for _, line := range strings.Split(view, "\n") {
		if strings.Contains(line, needle) {
			return strings.TrimSpace(line)
		}
	}
	return ""
}
