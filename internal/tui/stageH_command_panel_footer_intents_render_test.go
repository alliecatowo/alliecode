package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageHCommandPanelFooterIntentsRender(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.commandPanel.activate(commands.InteractivePanel{
		Title: "panel",
		Items: []commands.InteractivePanelItem{{Label: "Row", Status: "item"}},
		FooterIntents: []types.RenderIntent{{
			Kind:    types.RenderIntentActionHints,
			Title:   "Actions",
			Summary: "Follow-up commands.",
			Hints:   []types.RenderActionHint{{Label: "Status", Command: "/status"}},
		}},
	}, "")
	view := app.renderCommandPanel()
	if !strings.Contains(view, "Actions") || !strings.Contains(view, "- Status: /status") {
		t.Fatalf("expected footer intents rendered, got %q", view)
	}
}
