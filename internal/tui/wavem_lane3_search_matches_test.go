package tui

import "testing"

func TestWaveMLane3_SearchMatchesForTextReuseLoweredInput(t *testing.T) {
	row := timelineEntry{kind: timelineAssistant, text: "Error alpha error beta", turn: 1}
	plain := timelineSearchText(row)
	matches := timelineSearchMatchesForText(row, plain, "error alpha error beta", "error", 7)
	if len(matches) != 2 {
		t.Fatalf("expected two matches, got %d", len(matches))
	}
	if matches[0].lineOffset != 7 || matches[1].occurrence != 1 {
		t.Fatalf("unexpected match metadata: %#v", matches)
	}
}
