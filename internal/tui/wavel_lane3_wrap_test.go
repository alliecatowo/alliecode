package tui

import "testing"

func TestWaveLLane3_WrapDisplayWidthTrimsLeadingWhitespaceOnContinuation(t *testing.T) {
	rows := wrapDisplayWidth("      alpha beta", 5)
	if len(rows) < 2 {
		t.Fatalf("expected wrapped rows, got %#v", rows)
	}
	if rows[0] == "" {
		t.Fatalf("expected first wrapped row to keep visible content, got %#v", rows)
	}
}

func TestWaveLLane3_RenderTimelineHandlesUltraNarrowWidth(t *testing.T) {
	rows := []timelineEntry{{kind: timelineTool, toolName: "bash", toolSummary: "command=go test ./...", toolInputPreview: "go test ./...", toolInputBytes: 14, toolState: toolProgressRunning, turn: 1}}
	content, visible, offsets, _, total := renderTimeline(rows, "", 2)
	if content == "" {
		t.Fatalf("expected non-empty content at narrow width")
	}
	if len(visible) != 1 || len(offsets) != 1 {
		t.Fatalf("expected one visible row with one offset, got visible=%v offsets=%v", visible, offsets)
	}
	if total < 1 {
		t.Fatalf("expected at least one line, got %d", total)
	}
}
