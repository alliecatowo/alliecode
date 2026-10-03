package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func TestCommandPanelPreviewFallbackUsesRawAssistantText(t *testing.T) {
	app := readySizedApp(t, 140, 24)
	app.commandPanel.activate(commands.InteractivePanel{
		Title: "panel",
		Items: []commands.InteractivePanelItem{{
			Label:   "Login",
			Status:  "item",
			Preview: []string{"LOGIN_STATUS", "provider=openai", "provider_ready=true", "account=dev", "logged_in=true", "login_count=1", "logout_count=0"},
		}},
	}, "")
	view := app.renderCommandPanel()
	if !strings.Contains(view, "LOGIN_STATUS") || !strings.Contains(view, "provider=openai") {
		t.Fatalf("expected raw preview contract text when preview intents are absent, got %q", view)
	}
}

func TestCommandPanelPreviewFallbackKeepsNarrativeRawText(t *testing.T) {
	app := readySizedApp(t, 140, 24)
	app.commandPanel.activate(commands.InteractivePanel{
		Title: "panel",
		Items: []commands.InteractivePanelItem{{
			Label:   "Narrative",
			Status:  "item",
			Preview: []string{"Deployment finished", "All checks passed"},
		}},
	}, "")
	view := app.renderCommandPanel()
	if !strings.Contains(view, "Deployment finished") || !strings.Contains(view, "All checks passed") {
		t.Fatalf("expected narrative preview raw text to remain visible, got %q", view)
	}
}
