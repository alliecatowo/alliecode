package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelPickerShiftTabMovesSelectionBackward(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/model ")
	if !app.modelPickerActive() {
		t.Fatalf("expected model picker active")
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyDown}, "down")
	selected := app.slashAutocomplete.modelPicker.selected
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if app.slashAutocomplete.modelPicker.selected >= selected {
		t.Fatalf("expected shift+tab to move selection backward, got %d from %d", app.slashAutocomplete.modelPicker.selected, selected)
	}
}

func TestModelPickerLeftArrowDismisses(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/model ")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyLeft}, "left")
	if app.modelPickerActive() {
		t.Fatalf("expected left arrow to dismiss model picker")
	}
}

func TestModelPickerTabThenShiftTabReverseKeepsPanelOpen(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/model ")
	if !app.modelPickerActive() {
		t.Fatalf("expected model picker active")
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyTab}, "tab")
	selected := app.slashAutocomplete.modelPicker.selected
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if !app.modelPickerActive() {
		t.Fatalf("expected reverse cycle to keep model picker open")
	}
	if app.inputMode != inputModeModelPicker {
		t.Fatalf("expected model-picker input mode, got %s", app.inputMode)
	}
	if app.slashAutocomplete.modelPicker.selected == selected {
		t.Fatalf("expected reverse cycle to move selection, stayed at %d", app.slashAutocomplete.modelPicker.selected)
	}
	plain := stripANSIForTest(app.View())
	if !strings.Contains(plain, "model picker: /model") {
		t.Fatalf("expected model picker surface rendered, got:\n%s", plain)
	}
}

func TestModelPickerOneEnterStageThenEnterApplies(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/model")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if !app.modelPickerActive() {
		t.Fatalf("expected first enter on /model to open model picker")
	}
	app = sendTestKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if app.modelPickerActive() {
		t.Fatalf("expected second enter to apply selection and close picker")
	}
	if len(app.timeline) == 0 {
		t.Fatalf("expected model apply to append timeline output")
	}
	if !strings.Contains(app.timeline[len(app.timeline)-1].text, "Model set to") {
		t.Fatalf("expected model apply confirmation, got %q", app.timeline[len(app.timeline)-1].text)
	}
}

func TestModelPickerShiftTabSequenceVariantMovesSelectionBackward(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/model ")
	if !app.modelPickerActive() {
		t.Fatalf("expected model picker active")
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyDown}, "down")
	selected := app.slashAutocomplete.modelPicker.selected
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'', '[', 'Z'}}, "shift-tab-csi-z")
	if app.slashAutocomplete.modelPicker.selected >= selected {
		t.Fatalf("expected shift+tab sequence alias to move selection backward, got %d from %d", app.slashAutocomplete.modelPicker.selected, selected)
	}
}
