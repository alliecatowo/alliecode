package tui

import (
	"strings"
	"testing"
)

func TestWaveMLane3_RenderTimelineLargeTranscriptProducesAllRows(t *testing.T) {
	rows := make([]timelineEntry, 0, 1200)
	for i := 0; i < 1200; i++ {
		rows = append(rows, timelineEntry{kind: timelineAssistant, text: strings.Repeat("chunk ", 18), turn: i + 1})
	}
	_, visible, offsets, matches, total := renderTimeline(rows, "", 80)
	if len(visible) != len(rows) || len(offsets) != len(rows) {
		t.Fatalf("expected %d visible rows and offsets, got %d/%d", len(rows), len(visible), len(offsets))
	}
	if len(matches) != 0 {
		t.Fatalf("expected no matches with empty query, got %d", len(matches))
	}
	if total <= len(rows) {
		t.Fatalf("expected wrapped content to exceed row count, got total=%d", total)
	}
}
