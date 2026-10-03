package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/references"
)

func TestReferenceModeShiftTabMovesSelectionBackward(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.refAuto.active = true
	app.refAuto.suggestions = []references.Suggestion{{Path: "internal/tui/app.go"}, {Path: "internal/tui/input.go"}, {Path: "internal/tui/search_surface.go"}}
	app.refAuto.selected = 0
	app.syncInputMode()
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyDown}, "down")
	selected := app.refAuto.selected
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if app.refAuto.selected >= selected {
		t.Fatalf("expected shift+tab to move selection backward, got %d from %d", app.refAuto.selected, selected)
	}
}

func TestReferenceModeTabThenShiftTabReverseKeepsDrawerOpen(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "@READ")
	plain := stripANSIForTest(app.View())
	if !strings.Contains(plain, "references:") {
		t.Fatalf("expected reference drawer open, got:\n%s", plain)
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyTab}, "tab")
	selected := app.refAuto.selected
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if !app.refAuto.active {
		t.Fatalf("expected shift+tab reverse to keep reference drawer open")
	}
	if app.inputMode != inputModeReference {
		t.Fatalf("expected reference input mode after reverse, got %s", app.inputMode)
	}
	if app.refAuto.selected == selected {
		t.Fatalf("expected reverse navigation to change selection, stayed at %d", app.refAuto.selected)
	}
}

func TestReferenceModeShiftTabSequenceVariantMovesSelectionBackward(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.refAuto.active = true
	app.refAuto.suggestions = []references.Suggestion{{Path: "internal/tui/app.go"}, {Path: "internal/tui/input.go"}, {Path: "internal/tui/search_surface.go"}}
	app.refAuto.selected = 0
	app.syncInputMode()
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyDown}, "down")
	selected := app.refAuto.selected
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'', '[', 'Z'}}, "shift-tab-csi-z")
	if app.refAuto.selected >= selected {
		t.Fatalf("expected shift+tab sequence alias to move selection backward, got %d from %d", app.refAuto.selected, selected)
	}
}
