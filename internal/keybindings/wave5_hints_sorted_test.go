package keybindings

import "testing"

func TestWave5HintsForModeSortsSimpleBeforeChord(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+k ctrl+r", Action: "search:history", Modes: []Mode{ModeSearch}},
		{Combo: "ctrl+o", Action: "search:quick_open", Modes: []Mode{ModeSearch}},
	})
	hints := set.HintsForMode(ModeSearch, 4)
	if len(hints) < 2 {
		t.Fatalf("expected at least two hints, got %#v", hints)
	}
	if hints[0] != "ctrl+o -> search:quick_open" {
		t.Fatalf("expected single-step combo before chord hint, got %#v", hints)
	}
}
