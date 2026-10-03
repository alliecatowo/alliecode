package tui

import (
	"strings"
	"testing"
)

func TestStageHTimelinePlainTextVisibleWhenNotContract(t *testing.T) {
	rows := []timelineEntry{{kind: timelineAssistant, turn: 1, text: "Deployment complete. No action required."}}
	content, _, _, _, _ := renderTimeline(rows, "", 120)
	if !strings.Contains(content, "Deployment complete. No action required.") {
		t.Fatalf("expected plain assistant text visible, got %q", content)
	}
}
