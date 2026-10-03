package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStageHSearchTextExcludesContractWhenIntentPresent(t *testing.T) {
	row := timelineEntry{
		kind: timelineAssistant,
		text: "STATUS_REPORT\nprovider=openai",
		intents: []types.RenderIntent{{
			Kind:   types.RenderIntentSummaryCard,
			Title:  "Status",
			Fields: []types.RenderField{{Label: "Provider", Value: "openai"}},
		}},
	}
	got := timelineSearchText(row)
	if strings.Contains(got, "STATUS_REPORT") {
		t.Fatalf("expected contract token excluded from search text, got %q", got)
	}
	if !strings.Contains(got, "Status") || !strings.Contains(got, "Provider: openai") {
		t.Fatalf("expected intent text included in search corpus, got %q", got)
	}
}
