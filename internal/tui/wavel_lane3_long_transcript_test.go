package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestWaveLLane3_LongTranscriptKeepsAnchorAcrossOverlayToggle(t *testing.T) {
	app := readySizedApp(t, 120, 32)
	for i := 0; i < 180; i++ {
		app.addTimeline(timelineEntry{kind: timelineAssistant, text: strings.Repeat("segment ", 16), turn: i + 1})
	}
	app.followTail = false
	app.viewport.SetYOffset(40)
	before := app.viewport.YOffset

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}, "type:/")
	opened := app.viewport.YOffset
	if opened > before {
		t.Fatalf("expected anchored/clamped offset while slash drawer opens, before=%d open=%d", before, opened)
	}

	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, "esc")
	after := app.viewport.YOffset
	if diff := absInt(after - opened); diff > 2 {
		t.Fatalf("expected stable offset after overlay close, open=%d after=%d", opened, after)
	}
}

func TestWaveLLane3_LongTranscriptKeepsAnchorWhileStreaming(t *testing.T) {
	app := readySizedApp(t, 120, 32)
	for i := 0; i < 200; i++ {
		app.addTimeline(timelineEntry{kind: timelineAssistant, text: strings.Repeat("output ", 20), turn: i + 1})
	}
	app.followTail = false
	app.viewport.SetYOffset(55)
	before := app.viewport.YOffset

	app.streamBuf.WriteString(strings.Repeat("stream ", 48))
	app.refreshViewport()
	during := app.viewport.YOffset
	if during > before {
		t.Fatalf("expected anchored/clamped offset while stream block grows, before=%d during=%d", before, during)
	}

	app.streamBuf.WriteString(strings.Repeat("delta ", 48))
	app.refreshViewport()
	after := app.viewport.YOffset
	if diff := absInt(after - during); diff > 2 {
		t.Fatalf("expected stable offset across stream deltas, during=%d after=%d", during, after)
	}
}

func TestWaveLLane3_TimelineFrameCachesByVersionWidthAndQuery(t *testing.T) {
	app := New(Config{})
	app.width = 80
	app.setSearchQueryValue("")
	app.addTimeline(timelineEntry{kind: timelineAssistant, text: "first", turn: 1})

	_, _, _, _, _ = app.timelineFrame(80)
	if !app.cachedTimeline.ready {
		t.Fatalf("expected timeline cache to be populated")
	}
	firstVersion := app.cachedTimeline.version

	_, _, _, _, _ = app.timelineFrame(80)
	if app.cachedTimeline.version != firstVersion {
		t.Fatalf("expected cache reuse with unchanged frame inputs")
	}

	app.updateToolProgress("missing", "", toolProgressRunning)
	_, _, _, _, _ = app.timelineFrame(100)
	if app.cachedTimeline.width != 100 {
		t.Fatalf("expected width-sensitive cache entry, got width=%d", app.cachedTimeline.width)
	}

	app.setSearchQueryValue("first")
	app.setStateValue(stateSearch)
	app.setSearchModeValue(searchModeTimeline)
	_, _, _, _, _ = app.timelineFrame(100)
	if app.cachedTimeline.query != "first" {
		t.Fatalf("expected timeline-search query cache entry, got query=%q", app.cachedTimeline.query)
	}
}
