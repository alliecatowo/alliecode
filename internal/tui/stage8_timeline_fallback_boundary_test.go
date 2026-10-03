package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStage8TimelineFallbackBoundaryIntentsBypassFormatter(t *testing.T) {
	rows := []timelineEntry{{
		kind:    timelineAssistant,
		turn:    1,
		text:    "LOGIN_STATUS\nprovider=openai\nprovider_ready=true\naccount=dev\nlogged_in=true\nlogin_count=1\nlogout_count=0",
		intents: []types.RenderIntent{{Kind: types.RenderIntentSummaryCard, Title: "Login", Fields: []types.RenderField{{Label: "Authenticated", Value: "yes"}}}},
	}}
	content, _, _, _, _ := renderTimeline(rows, "", 120)
	if !strings.Contains(content, "Authenticated: yes") {
		t.Fatalf("expected intent body in timeline, got %q", content)
	}
	if strings.Contains(content, "provider=openai") {
		t.Fatalf("expected raw fallback suppressed when intents are present, got %q", content)
	}
}
