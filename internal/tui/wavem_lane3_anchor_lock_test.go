package tui

import "testing"

func TestWaveMLane3_CaptureAnchorLockHonorsScrolledState(t *testing.T) {
	app := readySizedApp(t, 120, 28)
	for i := 0; i < 80; i++ {
		app.addTimeline(timelineEntry{kind: timelineAssistant, text: "line line line line", turn: i + 1})
	}
	app.followTail = false
	app.viewport.SetYOffset(17)
	app.captureAnchorLock()
	if !app.anchorLockActive || app.anchorLockOffset != 17 {
		t.Fatalf("expected active anchor lock at offset 17, got active=%t offset=%d", app.anchorLockActive, app.anchorLockOffset)
	}
}
