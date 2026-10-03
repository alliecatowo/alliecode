package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func TestStageHPanelPreviewContractVisibleWithoutIntents(t *testing.T) {
	app := readySizedApp(t, 140, 24)
	app.commandPanel.activate(commands.InteractivePanel{Title: "panel", Items: []commands.InteractivePanelItem{{
		Label: "Status", Status: "item", Preview: []string{"LOGIN_STATUS", "provider=openai", "logged_in=true"},
	}}}, "")
	view := app.renderCommandPanel()
	if !strings.Contains(view, "LOGIN_STATUS") || !strings.Contains(view, "provider=openai") {
		t.Fatalf("expected panel preview contract visible without intents, got %q", view)
	}
}
