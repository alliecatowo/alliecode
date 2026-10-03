package tui

import "testing"

func TestStageHAssistantBoundaryKeepsPlainText(t *testing.T) {
	b := renderTimelineAssistantBody(timelineEntry{text: "Deployment finished successfully."})
	if !b.usedRawText || b.usedIntents {
		t.Fatalf("expected plain text boundary, got %#v", b)
	}
}
