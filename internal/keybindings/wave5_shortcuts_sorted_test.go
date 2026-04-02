package keybindings

import "testing"

func TestWave5ShortcutsForActionSortsByModeThenCombo(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+r", Action: "search:history", Modes: []Mode{ModeSearch}},
		{Combo: "ctrl+k ctrl+r", Action: "search:history", Modes: []Mode{ModeGlobal}},
		{Combo: "ctrl+alt+r", Action: "search:history", Modes: []Mode{ModeGlobal}},
	})
	shortcuts := set.ShortcutsForAction("search:history")
	if len(shortcuts) != 3 {
		t.Fatalf("expected 3 shortcuts, got %#v", shortcuts)
	}
	if shortcuts[0].Mode != ModeGlobal || shortcuts[0].Combo != "ctrl+alt+r" {
		t.Fatalf("expected sorted global combo first, got %#v", shortcuts)
	}
}
