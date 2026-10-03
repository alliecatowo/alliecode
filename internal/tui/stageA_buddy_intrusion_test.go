package tui

import (
	"testing"
	"time"
)

func TestStageABuddyIntrusionBubbleHiddenWithoutActiveReaction(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	app.buddy.bubble = ""
	app.buddy.bubbleUntil = time.Time{}

	if got := stripANSIForTest(app.renderBuddyBubbleLine()); got != " " {
		t.Fatalf("expected non-intrusive blank buddy bubble row, got %q", got)
	}
}
