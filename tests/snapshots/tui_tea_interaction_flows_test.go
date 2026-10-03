package snapshots_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTeaFlow_SlashDrawerOpenSelectEnterStagesCommand(t *testing.T) {
	app := readyApp(t, 140, 26)

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}, "type:/")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("mo")}, "type:mo")

	openView := stripANSI(app.View())
	if !strings.Contains(openView, "commands: /mo") {
		t.Fatalf("expected slash drawer to open with query, got:\n%s", openView)
	}
	if !strings.Contains(openView, "preview: /model") {
		t.Fatalf("expected /model preview in slash drawer, got:\n%s", openView)
	}

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	appliedView := stripANSI(app.View())
	if strings.Contains(appliedView, "commands: /") {
		t.Fatalf("expected slash drawer to close after enter, got:\n%s", appliedView)
	}
	if !strings.Contains(appliedView, "/model ") {
		t.Fatalf("expected /model to be staged in input after enter, got:\n%s", appliedView)
	}
}

func TestTeaFlow_ReferenceDrawerEnterAppliesSelection(t *testing.T) {
	app := readyApp(t, 140, 26)

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("@READ")}, "type:@READ")

	openView := stripANSI(app.View())
	if !strings.Contains(openView, "references:") {
		t.Fatalf("expected reference drawer to open, got:\n%s", openView)
	}
	if !strings.Contains(openView, "preview: @") {
		t.Fatalf("expected preview row in reference drawer, got:\n%s", openView)
	}

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	appliedView := stripANSI(app.View())
	if strings.Contains(appliedView, "references:") {
		t.Fatalf("expected reference drawer to close after enter, got:\n%s", appliedView)
	}
	if !strings.Contains(appliedView, "┃ @") {
		t.Fatalf("expected selected reference to be inserted into input line, got:\n%s", appliedView)
	}
}

func TestTeaFlow_ModelPickerPanelSelectApply(t *testing.T) {
	app := readyApp(t, 180, 28)

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyDown}, "down")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyDown}, "down")

	searchView := stripANSI(app.View())
	if !strings.Contains(searchView, "mode:quick-open") {
		t.Fatalf("expected quick-open panel to be active, got:\n%s", searchView)
	}
	if !strings.Contains(searchView, "Change model - Stage /model command in input") {
		t.Fatalf("expected model picker row to be selected, got:\n%s", searchView)
	}

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	appliedView := stripANSI(app.View())
	if strings.Contains(appliedView, "mode:quick-open") {
		t.Fatalf("expected quick-open panel to close after apply, got:\n%s", appliedView)
	}
	if !strings.Contains(appliedView, "hint: /model") {
		t.Fatalf("expected /model command hint after apply, got:\n%s", appliedView)
	}
	if !strings.Contains(appliedView, "/model ") {
		t.Fatalf("expected /model to be staged in input, got:\n%s", appliedView)
	}
}

func TestTeaFlow_EscapeClosesDrawersAndPicker(t *testing.T) {
	app := readyApp(t, 140, 26)

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/")}, "type:/")
	if !strings.Contains(stripANSI(app.View()), "commands: /") {
		t.Fatalf("expected slash drawer open before escape")
	}
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if strings.Contains(stripANSI(app.View()), "commands: /") {
		t.Fatalf("expected escape to close slash drawer")
	}

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("@READ")}, "type:@READ")
	if !strings.Contains(stripANSI(app.View()), "references:") {
		t.Fatalf("expected reference drawer open before escape")
	}
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if strings.Contains(stripANSI(app.View()), "references:") {
		t.Fatalf("expected escape to close reference drawer")
	}

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	if !strings.Contains(stripANSI(app.View()), "mode:quick-open") {
		t.Fatalf("expected quick-open active before escape")
	}
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	if strings.Contains(stripANSI(app.View()), "mode:quick-open") {
		t.Fatalf("expected escape to exit quick-open panel")
	}
}

func TestTeaFlow_QuickOpenOpenCloseKeepsComposerVisible(t *testing.T) {
	app := readyApp(t, 120, 22)
	for i := 0; i < 6; i++ {
		app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("hello")}, "type:hello")
		app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	}

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	if !strings.Contains(stripANSI(app.View()), "mode:quick-open") {
		t.Fatalf("expected quick-open open state")
	}
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	plain := stripANSI(app.View())
	if strings.Contains(plain, "mode:quick-open") {
		t.Fatalf("expected quick-open closed state")
	}
	if !strings.Contains(plain, "input mode: chat") || !strings.Contains(plain, "┃ Type a message") {
		t.Fatalf("expected composer stack visible after open/close, got:\n%s", plain)
	}
}

func TestTeaFlow_OneEnterSlashModelToPickerToApply(t *testing.T) {
	app := readyApp(t, 180, 28)

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/model")}, "type:/model")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	staged := stripANSI(app.View())
	if !strings.Contains(staged, "model picker: /model") {
		t.Fatalf("expected one-enter /model to open picker, got:\n%s", staged)
	}
	if !strings.Contains(staged, "┃ /model ") {
		t.Fatalf("expected /model staging with trailing space, got:\n%s", staged)
	}

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	applied := stripANSI(app.View())
	if strings.Contains(applied, "model picker: /model") {
		t.Fatalf("expected picker closed after second enter apply, got:\n%s", applied)
	}
	if !strings.Contains(applied, "Model set to") {
		t.Fatalf("expected model apply confirmation after one-enter picker apply, got:\n%s", applied)
	}
}

func TestTeaFlow_OneEnterSlashStatusRunsImmediately(t *testing.T) {
	app := readyApp(t, 180, 28)

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/status")}, "type:/status")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	plain := stripANSI(app.View())
	if strings.Contains(plain, "commands: /") {
		t.Fatalf("expected slash drawer closed after one-enter run, got:\n%s", plain)
	}
	if !strings.Contains(plain, "Provider:") || !strings.Contains(plain, "Tasks:") {
		t.Fatalf("expected status output rendered after one-enter run, got:\n%s", plain)
	}
}

func TestTeaFlow_OneEnterQuickOpenHistoryAppliesSelection(t *testing.T) {
	app := readyApp(t, 180, 28)
	for _, text := range []string{"deploy release", "open logs"} {
		app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(text)}, "type:history")
		app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	}

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlR}, "ctrl+r")
	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("log")}, "type:log")
	before := stripANSI(app.View())
	if !strings.Contains(before, "mode:history") {
		t.Fatalf("expected history search mode before apply, got:\n%s", before)
	}

	app = sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	after := stripANSI(app.View())
	if strings.Contains(after, "mode:history") {
		t.Fatalf("expected history mode closed after one-enter apply, got:\n%s", after)
	}
	if !strings.Contains(after, "┃ open logs") {
		t.Fatalf("expected selected history prompt staged into input, got:\n%s", after)
	}
}
