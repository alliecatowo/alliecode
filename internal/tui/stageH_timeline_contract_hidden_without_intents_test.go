package tui

import (
	"strings"
	"testing"
)

func TestStageHTimelineContractVisibleWithoutIntents(t *testing.T) {
	rows := []timelineEntry{{kind: timelineAssistant, turn: 1, text: "FILES_STATUS\ncount=2\nadds=1"}}
	content, _, _, _, _ := renderTimeline(rows, "", 120)
	if !strings.Contains(content, "FILES_STATUS") || !strings.Contains(content, "count=2") {
		t.Fatalf("expected contract payload visible without intents, got %q", content)
	}
}
