package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRenderTimelineFallsBackToRawAssistantTextWithoutIntents(t *testing.T) {
	rows := []timelineEntry{{
		kind: timelineAssistant,
		text: "LOGIN_STATUS\nlogged_in=true\nprovider=openai\naccount=dev@acme\nprovider_ready=true\nlogin_count=2\nlogout_count=1",
		turn: 1,
	}}
	content, _, _, _, _ := renderTimeline(rows, "", 120)
	if !strings.Contains(content, "[01] AI") {
		t.Fatalf("expected assistant row shell to remain, got %q", content)
	}
	if !containsAllTokens(content, []string{"LOGIN_STATUS", "provider=openai", "logged_in=true"}) {
		t.Fatalf("expected structured contract tokens visible in raw fallback, got %q", content)
	}
}

func TestRenderTimelinePrefersStructuredIntentsOverRawFallback(t *testing.T) {
	rows := []timelineEntry{{
		kind: timelineAssistant,
		text: "LOGIN_STATUS\nlogged_in=true\nprovider=openai\naccount=dev@acme\nprovider_ready=true\nlogin_count=2\nlogout_count=1",
		turn: 1,
		intents: []types.RenderIntent{{
			Kind:       types.RenderIntentDetailRows,
			Title:      "Login",
			DetailRows: []types.RenderDetailRow{{Label: "Authenticated", Value: "yes", Status: "ok", Detail: "Provider state from typed render intent"}},
		}},
	}}
	content, _, _, _, _ := renderTimeline(rows, "", 120)
	if !containsAllTokens(content, []string{"Login", "Authenticated: yes [ok] - Provider state from typed render intent"}) {
		t.Fatalf("expected structured intent content, got %q", content)
	}
	if strings.Contains(content, "provider=openai") {
		t.Fatalf("expected raw text to be bypassed when intents exist, got %q", content)
	}
}

func TestRenderTimelineKeepsRawNarrativeAssistantTextWithoutIntents(t *testing.T) {
	rows := []timelineEntry{{
		kind: timelineAssistant,
		text: "Deployment completed successfully. No pending actions.",
		turn: 1,
	}}
	content, _, _, _, _ := renderTimeline(rows, "", 120)
	if !containsAllTokens(content, []string{"Deployment completed successfully", "No pending actions"}) {
		t.Fatalf("expected plain narrative assistant text, got %q", content)
	}
}

func containsAllTokens(s string, wants []string) bool {
	for _, want := range wants {
		if !strings.Contains(s, want) {
			return false
		}
	}
	return true
}
