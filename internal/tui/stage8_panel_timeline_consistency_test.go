package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStage8PanelTimelineConsistencyForDetailRowsIntent(t *testing.T) {
	intents := []types.RenderIntent{{Kind: types.RenderIntentDetailRows, Title: "Runtime source", DetailRows: []types.RenderDetailRow{{Label: "Provider", Value: "anthropic", Status: "ready", Detail: "resolved at submit"}}}}
	app := readySizedApp(t, 160, 28)
	app.commandPanel.activate(commands.InteractivePanel{Title: "panel", Items: []commands.InteractivePanelItem{{Label: "row", Status: "item", PreviewIntents: intents}}}, "")
	panel := app.renderCommandPanel()
	timeline, _, _, _, _ := renderTimeline([]timelineEntry{{kind: timelineAssistant, turn: 1, intents: intents}}, "", 120)
	for _, token := range []string{"Runtime source", "Provider: anthropic [ready] - resolved at submit"} {
		if !strings.Contains(panel, token) {
			t.Fatalf("expected panel to include %q, got %q", token, panel)
		}
		if !strings.Contains(timeline, token) {
			t.Fatalf("expected timeline to include %q, got %q", token, timeline)
		}
	}
}
