package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func TestStageHPanelPreviewPlainTextVisible(t *testing.T) {
	app := readySizedApp(t, 140, 24)
	app.commandPanel.activate(commands.InteractivePanel{Title: "panel", Items: []commands.InteractivePanelItem{{
		Label: "Narrative", Status: "item", Preview: []string{"Deployment finished", "All checks passed"},
	}}}, "")
	view := app.renderCommandPanel()
	if !strings.Contains(view, "Deployment finished") || !strings.Contains(view, "All checks passed") {
		t.Fatalf("expected panel plain preview visible, got %q", view)
	}
}
