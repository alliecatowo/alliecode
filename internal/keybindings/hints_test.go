package keybindings

import "testing"

func TestActionsForModeOmitsUnbound(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+f", Action: "search", Modes: []Mode{ModeSearch}},
		{Combo: "ctrl+x", Action: "", Modes: []Mode{ModeSearch}, Unbind: true},
	})
	actions := set.ActionsForMode(ModeSearch)
	if len(actions) != 1 || actions[0] != "search" {
		t.Fatalf("unexpected actions list: %#v", actions)
	}
}
