package tui

import (
	"strings"
	"testing"
)

func TestStageBNarrowWidthToolRowRemainsReadable(t *testing.T) {
	rows := []timelineEntry{{
		kind:             timelineTool,
		turn:             2,
		toolName:         "read",
		toolState:        toolProgressDone,
		toolSummary:      "filePath=/very/long/path/to/some/file/that/needs/truncation.txt | offset=1200 | limit=200",
		toolInputPreview: "line1 line2 line3 line4 line5",
		toolInputBytes:   512,
	}}
	rendered, _, _, _, _ := renderTimeline(rows, "", 28)
	for _, token := range []string{"TOOL read [DONE]", "input:", "output: 512 chars"} {
		if !strings.Contains(rendered, token) {
			t.Fatalf("expected %q in narrow render, got %q", token, rendered)
		}
	}
}
