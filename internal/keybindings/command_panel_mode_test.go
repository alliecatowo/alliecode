package keybindings

import "testing"

func TestCommandPanelModeStackIncludesAutocompleteFallback(t *testing.T) {
	got := InputModeStack(ModeCommandPanel)
	want := []Mode{ModeCommandPanel, ModeAutocomplete, ModeChat, ModeGlobal}
	if len(got) != len(want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %#v want %#v", got, want)
		}
	}
}
