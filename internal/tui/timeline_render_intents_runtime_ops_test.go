package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRenderTimelineRuntimeOpsIntentCards(t *testing.T) {
	rows := []timelineEntry{{
		kind: timelineAssistant,
		turn: 3,
		text: "BRANCH_STATUS\nactive=main\ncount=2\ncreated=1\nswitches=4",
		intents: []types.RenderIntent{{
			Kind:    types.RenderIntentSummaryCard,
			Title:   "Branch status",
			Summary: "Current branch and mutation counters.",
			Fields: []types.RenderField{
				{Label: "Active", Value: "main"},
				{Label: "Branches", Value: "2"},
			},
		}},
	}}
	content, _, _, _, _ := renderTimeline(rows, "", 120)
	if !strings.Contains(content, "Branch status") || !strings.Contains(content, "Active: main") {
		t.Fatalf("expected structured branch status content, got %q", content)
	}
	if strings.Contains(content, "BRANCH_STATUS") {
		t.Fatalf("expected no legacy branch header when intents exist, got %q", content)
	}
}

func TestRenderTimelineRuntimeOpsShowsStructuredContractWhenNoIntent(t *testing.T) {
	rows := []timelineEntry{{kind: timelineAssistant, turn: 1, text: "FILES_STATUS\ncount=2\nproject_paths=1\nadds=2\nremoves=0\nclears=0\nlast_action=add\nlast_path=README.md"}}
	content, _, _, _, _ := renderTimeline(rows, "", 120)
	if !strings.Contains(content, "[01] AI") {
		t.Fatalf("expected assistant row shell to remain, got %q", content)
	}
	if !strings.Contains(content, "FILES_STATUS") {
		t.Fatalf("expected structured contract text visible without intents, got %q", content)
	}
}
