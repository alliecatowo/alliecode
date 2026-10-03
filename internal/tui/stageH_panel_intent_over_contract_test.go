package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageHPanelIntentPreviewOverridesContractText(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.commandPanel.activate(commands.InteractivePanel{Title: "panel", Items: []commands.InteractivePanelItem{{
		Label:  "Status",
		Status: "item",
		PreviewIntents: []types.RenderIntent{{
			Kind:       types.RenderIntentDetailRows,
			Title:      "Auth status",
			DetailRows: []types.RenderDetailRow{{Label: "Provider", Value: "openai", Status: "ready"}},
		}},
		Preview: []string{"LOGIN_STATUS", "provider=openai", "provider_ready=true"},
	}}}, "")
	view := app.renderCommandPanel()
	if !strings.Contains(view, "Auth status") || !strings.Contains(view, "Provider: openai [ready]") {
		t.Fatalf("expected intent preview content, got %q", view)
	}
	if strings.Contains(view, "LOGIN_STATUS") {
		t.Fatalf("expected contract preview hidden when intents exist, got %q", view)
	}
}
