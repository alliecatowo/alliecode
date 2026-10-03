package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRenderTimelineAssistantBodyUsesIntentsFirst(t *testing.T) {
	b := renderTimelineAssistantBody(timelineEntry{
		text:    "STATUS_REPORT\nprovider=openai",
		intents: []types.RenderIntent{{Kind: types.RenderIntentSummaryCard, Title: "Status", Fields: []types.RenderField{{Label: "Provider", Value: "openai"}}}},
	})
	if !b.usedIntents || b.usedRawText {
		t.Fatalf("expected intent-first boundary, got %#v", b)
	}
	if !strings.Contains(b.body, "Status") || !strings.Contains(b.body, "Provider: openai") {
		t.Fatalf("expected intent render body, got %q", b.body)
	}
}

func TestRenderTimelineAssistantBodyKeepsRawNarrativeText(t *testing.T) {
	b := renderTimelineAssistantBody(timelineEntry{text: "Deployment finished successfully. 2 services restarted."})
	if !b.usedRawText {
		t.Fatalf("expected plain narrative text path, got %#v", b)
	}
	if !strings.Contains(b.body, "Deployment finished successfully") {
		t.Fatalf("expected narrative body, got %q", b.body)
	}
}
