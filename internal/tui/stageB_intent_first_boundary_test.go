package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageBIntentFirstSuppressesFallbackFormattingBody(t *testing.T) {
	b := renderTimelineAssistantBody(timelineEntry{
		text: "LOGIN_STATUS\nprovider=openai\nprovider_ready=true",
		intents: []types.RenderIntent{{
			Kind:       types.RenderIntentDetailRows,
			Title:      "Auth status",
			DetailRows: []types.RenderDetailRow{{Label: "Provider", Value: "openai", Status: "ready"}},
		}},
	})
	if !b.usedIntents {
		t.Fatalf("expected intents path, got %#v", b)
	}
	if strings.Contains(b.body, "Login status") {
		t.Fatalf("expected fallback body suppressed when intents exist, got %q", b.body)
	}
	if !strings.Contains(b.body, "Auth status") {
		t.Fatalf("expected structured intent body, got %q", b.body)
	}
}
