package tui

import "testing"

func TestWave5HistoryDetailsShowsSixRowsBeforeOverflow(t *testing.T) {
	entry := historySearchEntry{text: "one\ntwo\nthree\nfour\nfive\nsix\nseven\neight"}
	state := newHistorySearchState([]historySearchEntry{entry})
	lines := state.selectedDetailLines(120)
	if len(lines) != 9 {
		t.Fatalf("expected heading + selected row + 6 rows + overflow line, got %#v", lines)
	}
	if lines[len(lines)-1] != "  ... +2 more lines" {
		t.Fatalf("expected overflow line, got %q", lines[len(lines)-1])
	}
}
