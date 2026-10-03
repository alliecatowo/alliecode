package tui

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/references"
)

func TestStage5PanelStackPriorityCommandPanelOverlaysSlashAndReference(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.state = stateIdle
	app.commandPanel.active = true
	app.commandPanel.panel = commands.InteractivePanel{Command: "permissions"}
	app.commandPanel.selected = 0
	app.slashAutocomplete.items = []commands.Suggestion{{Name: "status"}}
	app.refAuto.active = true
	app.refAuto.suggestions = []references.Suggestion{{Path: "internal/tui/app.go"}}
	app.syncInputMode()

	panels := app.inputPanels()
	if len(panels) == 0 || panels[0].key != "command-panel" {
		t.Fatalf("expected command panel to own input-adjacent stack, got %#v", panels)
	}
}

func TestStage5PanelStackPriorityModelPickerOverSlashAndReference(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.input.SetValue("/model ")
	app.syncSlashAutocomplete()
	app.syncModelPicker()
	app.refAuto.active = true
	app.refAuto.suggestions = []references.Suggestion{{Path: "internal/tui/input.go"}}
	app.syncInputMode()

	panels := app.inputPanels()
	if len(panels) == 0 || panels[0].key != "model-picker" {
		t.Fatalf("expected model picker to own input-adjacent stack, got %#v", panels)
	}
}

func TestStage5PanelStackPriorityPermissionPromptOverlaysAllPanels(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.state = statePermissionPrompt
	app.commandPanel.active = true
	app.slashAutocomplete.items = []commands.Suggestion{{Name: "status"}}
	app.refAuto.active = true
	app.refAuto.suggestions = []references.Suggestion{{Path: "internal/tui/app.go"}}

	panels := app.activityPanels()
	if len(panels) != 1 || panels[0].key != "permission" {
		t.Fatalf("expected permission prompt to own activity stack, got %#v", panels)
	}
}
