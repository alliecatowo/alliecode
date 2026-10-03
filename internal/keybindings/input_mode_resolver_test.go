package keybindings

import "testing"

func TestResolveActionForInputModeUsesStackPrecedence(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+o", Action: "global-open", Modes: []Mode{ModeGlobal}},
		{Combo: "ctrl+o", Action: "search-open", Modes: []Mode{ModeSearch}},
		{Combo: "ctrl+o", Action: "quick-open", Modes: []Mode{ModeQuickOpen}},
	})

	binding, ok := set.ResolveActionForInputMode(ModeQuickOpen, "ctrl+o")
	if !ok {
		t.Fatalf("expected combo to resolve")
	}
	if binding.Action != "quick-open" {
		t.Fatalf("expected quick-open layer to win, got %q", binding.Action)
	}
}

func TestResolveActionForInputModeHonorsUnbind(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+r", Action: "global-history", Modes: []Mode{ModeGlobal}},
		{Combo: "ctrl+r", Action: "", Unbind: true, Modes: []Mode{ModeHistory}},
	})

	if _, ok := set.ResolveActionForInputMode(ModeHistory, "ctrl+r"); ok {
		t.Fatalf("expected history layer unbind to block fallback")
	}
}
