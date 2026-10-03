package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestCommandPanelOpenRuntimeOpsCommands(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	for _, name := range []string{"branch", "diff", "files", "memory", "theme", "output-style", "privacy-settings", "upgrade", "resume", "plan"} {
		if !app.openInteractiveCommandPanel(name) {
			t.Fatalf("expected interactive panel to open for %q", name)
		}
		view := app.renderCommandPanel()
		if !strings.Contains(strings.ToLower(view), name) {
			t.Fatalf("expected panel view to mention %q, got %q", name, view)
		}
		app.closeInteractiveCommandPanel()
	}
}

func TestCommandPanelRuntimeOpsPreviewIntentsRender(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.commandPanel.activate(commands.InteractivePanel{
		Command:       "upgrade",
		Title:         "upgrade panel: /upgrade",
		HeaderIntents: []types.RenderIntent{{Kind: types.RenderIntentSummaryCard, Title: "Upgrade status", Fields: []types.RenderField{{Label: "Requested", Value: "no"}}}},
		Items: []commands.InteractivePanelItem{{
			Key:        "pro",
			Section:    "Plans",
			Label:      "Request pro plan",
			Detail:     "Queue pro upgrade guidance",
			Status:     "plan",
			ApplyInput: "/upgrade pro",
			ApplyMode:  commands.PanelApplySubmit,
			PreviewIntents: []types.RenderIntent{{
				Kind:    types.RenderIntentSummaryCard,
				Title:   "Upgrade requested",
				Summary: "Requested provider upgrade guidance.",
				Fields:  []types.RenderField{{Label: "Plan", Value: "pro"}},
			}},
		}},
	}, "/upgrade")
	view := app.renderCommandPanel()
	for _, token := range []string{"Upgrade status", "Request pro plan", "Upgrade requested", "Plan: pro"} {
		if !strings.Contains(view, token) {
			t.Fatalf("expected token %q in panel render, got %q", token, view)
		}
	}
}
