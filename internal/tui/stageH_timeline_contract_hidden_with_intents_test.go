package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageHTimelineContractHiddenWhenIntentExists(t *testing.T) {
	rows := []timelineEntry{{
		kind: timelineAssistant,
		turn: 2,
		text: "BRANCH_STATUS\nactive=main\ncount=2",
		intents: []types.RenderIntent{{
			Kind:   types.RenderIntentSummaryCard,
			Title:  "Branch status",
			Fields: []types.RenderField{{Label: "Active", Value: "main"}},
		}},
	}}
	content, _, _, _, _ := renderTimeline(rows, "", 120)
	if strings.Contains(content, "BRANCH_STATUS") || strings.Contains(content, "active=main") {
		t.Fatalf("expected contract payload hidden when intents exist, got %q", content)
	}
}
