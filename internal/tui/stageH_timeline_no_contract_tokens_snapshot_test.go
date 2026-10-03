package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageHTimelineNoContractTokensSnapshot(t *testing.T) {
	rows := []timelineEntry{
		{kind: timelineAssistant, turn: 1, text: "STATUS_REPORT\nprovider=openai\nprovider_ready=true", intents: []types.RenderIntent{{Kind: types.RenderIntentSummaryCard, Title: "Status", Fields: []types.RenderField{{Label: "Provider", Value: "openai"}}}}},
		{kind: timelineAssistant, turn: 2, text: "FILES_STATUS\ncount=2\nadds=1"},
		{kind: timelineAssistant, turn: 3, text: "Narrative text remains visible."},
	}
	content, _, _, _, _ := renderTimeline(rows, "", 120)
	if strings.Contains(content, "STATUS_REPORT") {
		t.Fatalf("expected intent-backed contract tokens suppressed from timeline snapshot, got %q", content)
	}
	if !strings.Contains(content, "FILES_STATUS") {
		t.Fatalf("expected non-intent raw contract row to remain visible, got %q", content)
	}
	if !strings.Contains(content, "Narrative text remains visible.") || !strings.Contains(content, "Status") {
		t.Fatalf("expected intent and narrative content in snapshot, got %q", content)
	}
}
