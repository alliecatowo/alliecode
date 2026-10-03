package keybindings

import "testing"

func TestCompactHintsForModesFormatsActions(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+o", Action: "search:quick_open", Modes: []Mode{ModeQuickOpen}},
		{Combo: "enter", Action: "confirm:accept", Modes: []Mode{ModeQuickOpen}},
	})
	hints := set.CompactHintsForModes([]Mode{ModeQuickOpen}, 4)
	if len(hints) != 2 {
		t.Fatalf("expected two hints, got %#v", hints)
	}
	if hints[0] != "ctrl+o search/quick_open" {
		t.Fatalf("unexpected first compact hint: %#v", hints)
	}
}
