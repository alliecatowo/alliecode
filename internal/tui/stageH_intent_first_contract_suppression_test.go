package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageHIntentFirstTimelineSuppressesLegacyContractWhenIntentsExist(t *testing.T) {
	rows := []timelineEntry{{
		kind: timelineAssistant,
		turn: 1,
		text: "STATUS_REPORT\nprovider=openai\nprovider_ready=true",
		intents: []types.RenderIntent{{
			Kind:   types.RenderIntentSummaryCard,
			Title:  "Status",
			Fields: []types.RenderField{{Label: "Provider", Value: "openai"}},
		}},
	}}
	content, _, _, _, _ := renderTimeline(rows, "", 120)
	if !strings.Contains(content, "Status") || !strings.Contains(content, "Provider: openai") {
		t.Fatalf("expected intent content rendered, got %q", content)
	}
	for _, forbidden := range []string{"STATUS_REPORT", "provider=openai", "provider_ready=true"} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("expected legacy contract token %q to be suppressed, got %q", forbidden, content)
		}
	}
}

func TestStageHIntentFirstTimelineKeepsRawContractWithoutIntents(t *testing.T) {
	rows := []timelineEntry{{
		kind: timelineAssistant,
		turn: 1,
		text: "FILES_STATUS\ncount=2\nproject_paths=1\nlast_action=add",
	}}
	content, _, _, _, _ := renderTimeline(rows, "", 120)
	for _, required := range []string{"FILES_STATUS", "count=2", "project_paths=1", "last_action=add"} {
		if !strings.Contains(content, required) {
			t.Fatalf("expected raw contract token %q to remain visible, got %q", required, content)
		}
	}
}

func TestStageHIntentFirstPanelKeepsRawPreviewContractWithoutIntents(t *testing.T) {
	app := readySizedApp(t, 140, 24)
	app.commandPanel.activate(commands.InteractivePanel{
		Title: "panel",
		Items: []commands.InteractivePanelItem{{
			Label:   "Status",
			Status:  "item",
			Preview: []string{"LOGIN_STATUS", "provider=openai", "logged_in=true"},
		}},
	}, "")
	view := app.renderCommandPanel()
	for _, required := range []string{"LOGIN_STATUS", "provider=openai", "logged_in=true"} {
		if !strings.Contains(view, required) {
			t.Fatalf("expected raw preview contract token %q visible, got %q", required, view)
		}
	}
}

func TestStageHIntentFirstPanelUsesPreviewIntentsWhenPresent(t *testing.T) {
	app := readySizedApp(t, 140, 24)
	app.commandPanel.activate(commands.InteractivePanel{
		Title: "panel",
		Items: []commands.InteractivePanelItem{{
			Label:  "Status",
			Status: "item",
			PreviewIntents: []types.RenderIntent{{
				Kind:  types.RenderIntentDetailRows,
				Title: "Auth status",
				DetailRows: []types.RenderDetailRow{{
					Label:  "Provider",
					Value:  "openai",
					Status: "ready",
				}},
			}},
			Preview: []string{"LOGIN_STATUS", "provider=openai", "provider_ready=true"},
		}},
	}, "")
	view := app.renderCommandPanel()
	if !strings.Contains(view, "Auth status") || !strings.Contains(view, "Provider: openai [ready]") {
		t.Fatalf("expected intent preview rendered, got %q", view)
	}
	if strings.Contains(view, "LOGIN_STATUS") {
		t.Fatalf("expected legacy preview contract suppressed when intents exist, got %q", view)
	}
}

func TestStageHIntentFirstTimelineKeepsNarrativeAssistantText(t *testing.T) {
	rows := []timelineEntry{{
		kind: timelineAssistant,
		turn: 1,
		text: "Deployment complete. Services are healthy.",
	}}
	content, _, _, _, _ := renderTimeline(rows, "", 120)
	if !strings.Contains(content, "Deployment complete. Services are healthy.") {
		t.Fatalf("expected narrative assistant text to remain visible, got %q", content)
	}
}
