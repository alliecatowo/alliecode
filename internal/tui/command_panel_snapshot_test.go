package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestCommandPanelSnapshotRenderIncludesPreviewAndFooter(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.commandPanel.activate(commands.InteractivePanel{
		Command:       "history",
		Title:         "history panel: /history",
		Subtitle:      "entries=1 views=0",
		HeaderIntents: []types.RenderIntent{{Kind: types.RenderIntentSummaryCard, Title: "History status", Fields: []types.RenderField{{Label: "Entries", Value: "1"}}}},
		Items: []commands.InteractivePanelItem{{
			Key:            "h-1",
			Section:        "Entries",
			Label:          "Fix deploy (h-1)",
			Detail:         "model=gpt-4o-mini turns=4 created=2026-04-01",
			Status:         "entry",
			ApplyInput:     "/history show h-1",
			PreviewIntents: []types.RenderIntent{{Kind: types.RenderIntentDetailRows, Title: "History entry", DetailRows: []types.RenderDetailRow{{Label: "Path", Value: "/tmp/h-1.json", Status: "stored", Detail: "Summary: repaired deployment lane"}}}},
			Preview:        []string{"Path: /tmp/h-1.json", "Summary: repaired deployment lane"},
		}},
	}, "/history")
	app.syncInputMode()

	view := app.renderCommandPanel()
	for _, needle := range []string{"drawer: history panel: /history", "History status", "## Entries", "preview: Fix deploy (h-1)", "apply: /history show h-1", "Path: /tmp/h-1.json [stored] - Summary: repaired deployment lane", "enter apply", "shift+tab/up prev", "esc close"} {
		if !strings.Contains(view, needle) {
			t.Fatalf("expected command panel render to include %q, got %q", needle, view)
		}
	}
}
