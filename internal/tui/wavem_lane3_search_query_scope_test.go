package tui

import "testing"

func TestWaveMLane3_TimelineFrameIgnoresQueryOutsideTimelineSearch(t *testing.T) {
	app := New(Config{})
	app.width = 80
	app.addTimeline(timelineEntry{kind: timelineAssistant, text: "needle row", turn: 1})
	app.setSearchQueryValue("needle")
	app.setStateValue(stateSearch)
	app.setSearchModeValue(searchModeQuickOpen)

	_, visible, _, _, _ := app.timelineFrame(80)
	if len(visible) != 1 {
		t.Fatalf("expected timeline to stay unfiltered in quick-open mode, got visible=%d", len(visible))
	}
	if app.cachedTimeline.query != "" {
		t.Fatalf("expected cached timeline query to be empty outside timeline mode, got %q", app.cachedTimeline.query)
	}
}
