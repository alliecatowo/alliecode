package keybindings

import "testing"

func TestWave5ResolveWithChordHonorsUnboundOverMatch(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+x", Action: "global-action", Modes: []Mode{ModeGlobal}},
		{Combo: "ctrl+x", Action: "", Unbind: true, Modes: []Mode{ModeSearch}},
	})
	result := set.ResolveWithChordInModes([]Mode{ModeSearch, ModeGlobal}, nil, "ctrl+x")
	if result.Type != ResolveUnbound {
		t.Fatalf("expected unbound in active layer, got %+v", result)
	}
}
