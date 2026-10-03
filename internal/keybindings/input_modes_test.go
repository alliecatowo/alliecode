package keybindings

import "testing"

func TestInputModeStackIncludesSpecificAndFallbackModes(t *testing.T) {
	tests := []struct {
		mode Mode
		want []Mode
	}{
		{mode: ModeSlash, want: []Mode{ModeSlash, ModeAutocomplete, ModeChat, ModeGlobal}},
		{mode: ModeReference, want: []Mode{ModeReference, ModeAutocomplete, ModeChat, ModeGlobal}},
		{mode: ModeQuickOpen, want: []Mode{ModeQuickOpen, ModeSearch, ModeAutocomplete, ModeChat, ModeGlobal}},
		{mode: ModeHistory, want: []Mode{ModeHistory, ModeSearch, ModeAutocomplete, ModeChat, ModeGlobal}},
		{mode: ModeTimeline, want: []Mode{ModeTimeline, ModeSearch, ModeAutocomplete, ModeChat, ModeGlobal}},
		{mode: ModeModelPicker, want: []Mode{ModeModelPicker, ModeAutocomplete, ModeSlash, ModeChat, ModeGlobal}},
		{mode: ModeCommandPanel, want: []Mode{ModeCommandPanel, ModeAutocomplete, ModeChat, ModeGlobal}},
		{mode: ModePermission, want: []Mode{ModePermission, ModeGlobal}},
	}

	for _, tc := range tests {
		got := InputModeStack(tc.mode)
		if len(got) != len(tc.want) {
			t.Fatalf("%s: got %#v want %#v", tc.mode, got, tc.want)
		}
		for i := range tc.want {
			if got[i] != tc.want[i] {
				t.Fatalf("%s: got %#v want %#v", tc.mode, got, tc.want)
			}
		}
	}
}
