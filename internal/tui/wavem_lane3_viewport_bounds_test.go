package tui

import (
	"strings"
	"testing"
)

func TestWaveMLane3_ViewportOffsetClampedAfterLargeShrink(t *testing.T) {
	app := readySizedApp(t, 100, 26)
	for i := 0; i < 160; i++ {
		app.addTimeline(timelineEntry{kind: timelineAssistant, text: strings.Repeat("segment ", 10), turn: i + 1})
	}
	app.followTail = false
	app.viewport.SetYOffset(70)
	app.width = 80
	app.height = 14
	app.recalcLayout()
	if app.viewport.YOffset < 0 {
		t.Fatalf("expected non-negative offset after shrink, got %d", app.viewport.YOffset)
	}
}
