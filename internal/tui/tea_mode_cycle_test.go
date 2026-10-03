package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/references"
)

func TestTeaModeCycleSearchCanPivotAcrossModes(t *testing.T) {
	app := readySizedApp(t, 160, 28)

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = typeTestText(t, app, "history")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if app.state != stateSearch || app.searchMode != searchModeHistory {
		t.Fatalf("expected quick-open enter on history row to pivot to history mode, state=%d mode=%d", app.state, app.searchMode)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	if app.searchMode != searchModeQuickOpen {
		t.Fatalf("expected ctrl+o to pivot back to quick-open, got %d", app.searchMode)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, "ctrl+f")
	if app.searchMode != searchModeTimeline {
		t.Fatalf("expected ctrl+f to pivot to timeline search, got %d", app.searchMode)
	}
}

func TestTeaModeCycleShiftTabWrapsQuickOpenAndHistory(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.input.history = []string{"deploy release", "open logs", "tail service logs"}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if app.quickOpen.selected != len(app.quickOpen.visible)-1 {
		t.Fatalf("expected quick-open shift+tab to wrap to last visible row, got %d of %d", app.quickOpen.selected, len(app.quickOpen.visible))
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlR}, "ctrl+r")
	app = typeTestText(t, app, "log")
	if len(app.history.matches) < 2 {
		t.Fatalf("expected multiple history matches for reverse-wrap test")
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyDown}, "down")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if app.history.selected != 0 {
		t.Fatalf("expected history shift+tab to reverse from second row to first, got %d", app.history.selected)
	}
}

func TestTeaModeCycleShiftTabWrapsSlashAndReferenceDrawers(t *testing.T) {
	app := readySizedApp(t, 160, 28)

	app = typeTestText(t, app, "/")
	if len(app.slashAutocomplete.items) < 2 {
		t.Fatalf("expected multiple slash items for reverse-wrap test")
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if app.slashAutocomplete.selected != len(app.slashAutocomplete.items)-1 {
		t.Fatalf("expected slash shift+tab to wrap to last row, got %d of %d", app.slashAutocomplete.selected, len(app.slashAutocomplete.items))
	}

	app.slashAutocomplete.clear()
	app.refAuto.active = true
	app.refAuto.suggestions = []references.Suggestion{{Path: "internal/tui/app.go"}, {Path: "internal/tui/search_surface.go"}}
	app.refAuto.selected = 0
	app.syncInputMode()
	if len(app.refAuto.suggestions) < 2 {
		t.Fatalf("expected multiple reference suggestions for reverse-wrap test")
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if app.refAuto.selected != len(app.refAuto.suggestions)-1 {
		t.Fatalf("expected reference shift+tab to wrap to last row, got %d of %d", app.refAuto.selected, len(app.refAuto.suggestions))
	}
}

func TestTeaModeCycleShiftTabWrapsModelPicker(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/model ")
	if len(app.slashAutocomplete.modelPicker.items) < 2 {
		t.Fatalf("expected multiple model picker rows for reverse-wrap test")
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if app.slashAutocomplete.modelPicker.selected != len(app.slashAutocomplete.modelPicker.items)-1 {
		t.Fatalf("expected model picker shift+tab to wrap to last row, got %d of %d", app.slashAutocomplete.modelPicker.selected, len(app.slashAutocomplete.modelPicker.items))
	}
}

func TestTeaModeCycleCtrlBindingsPivotAcrossSearchModesWithoutDroppingSearchState(t *testing.T) {
	app := readySizedApp(t, 160, 28)

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, "ctrl+f")
	if app.inputMode != inputModeSearch {
		t.Fatalf("expected timeline search mode, got %s", app.inputMode)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	if app.searchMode != searchModeQuickOpen || app.inputMode != inputModeQuickOpen {
		t.Fatalf("expected quick-open pivot from timeline, mode=%d input=%s", app.searchMode, app.inputMode)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlR}, "ctrl+r")
	if app.searchMode != searchModeHistory || app.inputMode != inputModeHistorySearch {
		t.Fatalf("expected history pivot from quick-open, mode=%d input=%s", app.searchMode, app.inputMode)
	}
}

func TestTeaModeCycleShiftTabSequenceWrapsQuickOpen(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'', '[', '1', ';', '2', 'Z'}}, "shift-tab-csi-1-2-z")
	if app.quickOpen.selected != len(app.quickOpen.visible)-1 {
		t.Fatalf("expected shift+tab sequence alias to wrap to last visible row, got %d of %d", app.quickOpen.selected, len(app.quickOpen.visible))
	}
}
