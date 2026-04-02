package keybindings

import "testing"

func TestWave5ResolveActionInStackRespectsModeOrder(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+r", Action: "global-history", Modes: []Mode{ModeGlobal}},
		{Combo: "ctrl+r", Action: "search-history", Modes: []Mode{ModeSearch}},
	})
	binding, ok := set.ResolveActionInStack([]Mode{ModeSearch, ModeGlobal}, "ctrl+r")
	if !ok || binding.Action != "search-history" {
		t.Fatalf("expected search mode action to win, got ok=%t binding=%+v", ok, binding)
	}
}
