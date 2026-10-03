package tui

import (
	"fmt"
	"testing"
)

func TestWaveMLane3_RenderTimelineQueryFilterLargeCorpus(t *testing.T) {
	rows := make([]timelineEntry, 0, 600)
	for i := 0; i < 600; i++ {
		text := fmt.Sprintf("event-%03d", i)
		if i%25 == 0 {
			text += " needle"
		}
		rows = append(rows, timelineEntry{kind: timelineAssistant, text: text, turn: i + 1})
	}
	_, visible, _, matches, _ := renderTimeline(rows, "needle", 90)
	if len(visible) != 24 {
		t.Fatalf("expected 24 filtered rows, got %d", len(visible))
	}
	if len(matches) != 24 {
		t.Fatalf("expected one match per filtered row, got %d", len(matches))
	}
}
