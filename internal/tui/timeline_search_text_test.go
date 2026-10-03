package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTimelineSearchTextIncludesIntentBodyAndSuppressesContractFallback(t *testing.T) {
	row := timelineEntry{kind: timelineAssistant, text: "STATUS_REPORT\nprovider=openai", intents: []types.RenderIntent{{Kind: types.RenderIntentSummaryCard, Title: "Status", Fields: []types.RenderField{{Label: "Provider", Value: "openai"}}}}}
	got := timelineSearchText(row)
	for _, token := range []string{"Status", "Provider: openai"} {
		if !strings.Contains(got, token) {
			t.Fatalf("expected %q in search text: %q", token, got)
		}
	}
	if strings.Contains(got, "STATUS_REPORT") {
		t.Fatalf("expected contract header suppressed from intent-backed search text, got %q", got)
	}
}

func TestTimelineSearchTextKeepsNarrativeAssistantText(t *testing.T) {
	row := timelineEntry{kind: timelineAssistant, text: "Status ready and stable."}
	got := timelineSearchText(row)
	if !strings.Contains(got, "Status ready and stable.") {
		t.Fatalf("expected narrative text in search corpus, got %q", got)
	}
}
