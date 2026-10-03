package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/references"
)

func TestTeaPanelTransitionSlashQueryOpensPermissionsPanel(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	if !app.openInteractiveCommandPanel("permissions") {
		t.Fatalf("expected permissions command panel to open")
	}

	if !app.commandPanel.active {
		t.Fatalf("expected permissions command panel to open from slash query")
	}
	if app.commandPanel.panel.Command != "permissions" {
		t.Fatalf("expected permissions panel, got %q", app.commandPanel.panel.Command)
	}
	if app.inputMode != inputModeCommandPanel {
		t.Fatalf("expected command-panel input mode, got %s", app.inputMode)
	}
	plain := stripANSIForTest(app.View())
	for _, want := range []string{"permissions panel: /permissions", "hints: command panel", "context: command panel /permissions"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("expected permissions panel view to contain %q, got:\n%s", want, plain)
		}
	}
}

func TestTeaPanelTransitionCommandPanelShiftTabWrapsToLastItem(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	if !app.openInteractiveCommandPanel("permissions") {
		t.Fatalf("expected permissions panel active")
	}
	if !app.commandPanel.active {
		t.Fatalf("expected permissions panel active")
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if app.commandPanel.selected != len(app.commandPanel.panel.Items)-1 {
		t.Fatalf("expected shift+tab to wrap to last panel row, got %d of %d", app.commandPanel.selected, len(app.commandPanel.panel.Items))
	}
}

func TestTeaPanelTransitionCommandPanelEnterSubmitsDefaultSelection(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	if !app.openInteractiveCommandPanel("permissions") {
		t.Fatalf("expected permissions panel active")
	}
	if !app.commandPanel.active {
		t.Fatalf("expected permissions panel active")
	}

	app = sendTestKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	if app.commandPanel.active {
		t.Fatalf("expected command panel closed after apply")
	}
	if len(app.timeline) != 1 {
		t.Fatalf("expected one timeline row after panel submit, got %d", len(app.timeline))
	}
	if !strings.Contains(app.timeline[0].text, "PERMISSIONS_SUMMARY") {
		t.Fatalf("expected permissions summary output, got %q", app.timeline[0].text)
	}
}

func TestTeaPanelTransitionCommandPanelLeftClosesWithoutApplying(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	if !app.openInteractiveCommandPanel("permissions") {
		t.Fatalf("expected permissions panel active")
	}
	if !app.commandPanel.active {
		t.Fatalf("expected permissions panel active")
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyLeft}, "left")
	if app.commandPanel.active {
		t.Fatalf("expected left to close command panel")
	}
	if got := app.input.Value(); got != "/permissions" {
		t.Fatalf("expected command text preserved after left close, got %q", got)
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected no timeline output after left close, got %d rows", len(app.timeline))
	}
}


func TestTeaPanelTransitionModelPickerPrecedesSlashAndReferencePanels(t *testing.T) {
	app := readySizedApp(t, 180, 30)
	app = typeTestText(t, app, "/model ")
	if !app.modelPickerActive() {
		t.Fatalf("expected model picker active")
	}

	app.refAuto.active = true
	app.refAuto.suggestions = []references.Suggestion{{Path: "internal/tui/app.go"}}
	app.syncInputMode()

	plain := stripANSIForTest(app.View())
	if !strings.Contains(plain, "model picker: /model") {
		t.Fatalf("expected model picker panel visible, got:\n%s", plain)
	}
	if strings.Contains(plain, "commands: /") {
		t.Fatalf("expected slash panel hidden while model picker active, got:\n%s", plain)
	}
	if strings.Contains(plain, "references:") {
		t.Fatalf("expected reference panel hidden while model picker active, got:\n%s", plain)
	}
}
