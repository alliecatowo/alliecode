package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStage5IntentTimelinePrefersIntentDetailRowsOverHeaderFormatting(t *testing.T) {
	rows := []timelineEntry{{
		kind: timelineAssistant,
		text: strings.Join([]string{
			"STATUS_REPORT",
			"provider=anthropic",
			"model=claude-sonnet",
		}, "\n"),
		turn: 4,
		intents: []types.RenderIntent{{
			Kind:       types.RenderIntentDetailRows,
			Title:      "Runtime source of truth",
			DetailRows: []types.RenderDetailRow{{Label: "Provider", Value: "anthropic", Status: "ready", Detail: "submit path uses resolved runtime selection"}},
		}},
	}}

	content, _, _, _, _ := renderTimeline(rows, "", 120)
	if !strings.Contains(content, "Runtime source of truth") {
		t.Fatalf("expected typed intent heading in timeline content, got %q", content)
	}
	if strings.Contains(content, "Status") {
		t.Fatalf("expected command formatter header to be bypassed when intents are present, got %q", content)
	}
}
