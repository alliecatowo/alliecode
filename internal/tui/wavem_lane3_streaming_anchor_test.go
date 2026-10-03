package tui

import (
	"strings"
	"testing"
)

func TestWaveMLane3_StreamingGrowthKeepsManualOffsetAnchored(t *testing.T) {
	app := readySizedApp(t, 120, 30)
	for i := 0; i < 170; i++ {
		app.addTimeline(timelineEntry{kind: timelineAssistant, text: strings.Repeat("line ", 20), turn: i + 1})
	}
	app.followTail = false
	app.viewport.SetYOffset(41)
	baseline := app.viewport.YOffset
	app.streamBuf.WriteString(strings.Repeat("delta ", 32))
	app.refreshViewport()
	app.streamBuf.WriteString(strings.Repeat("delta ", 32))
	app.refreshViewport()
	if diff := absInt(app.viewport.YOffset - baseline); diff > 2 {
		t.Fatalf("expected near-stable offset while stream grows, baseline=%d now=%d", baseline, app.viewport.YOffset)
	}
}
