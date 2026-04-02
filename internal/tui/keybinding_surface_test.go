package tui

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/keybindings"
)

func TestDefaultKeybindingSetIncludesLayeredNavigation(t *testing.T) {
	set := defaultKeybindingSet()
	if set == nil {
		t.Fatalf("expected keybinding set")
	}

	required := []struct {
		combo  string
		action string
		mode   keybindings.Mode
	}{
		{combo: "ctrl+k ctrl+f", action: "search:timeline", mode: keybindings.ModeGlobal},
		{combo: "ctrl+k ctrl+o", action: "search:quick_open", mode: keybindings.ModeGlobal},
		{combo: "ctrl+k ctrl+r", action: "search:history", mode: keybindings.ModeGlobal},
		{combo: "home", action: "select:first", mode: keybindings.ModeSearch},
		{combo: "end", action: "select:last", mode: keybindings.ModeSearch},
		{combo: "pgdown", action: "select:page_down", mode: keybindings.ModeAutocomplete},
		{combo: "ctrl+n", action: "history:next", mode: keybindings.ModeInsert},
		{combo: "ctrl+p", action: "history:previous", mode: keybindings.ModeInsert},
		{combo: "ctrl+j", action: "input:newline", mode: keybindings.ModeInsert},
		{combo: "ctrl+w", action: "input:delete_word", mode: keybindings.ModeInsert},
		{combo: "ctrl+u", action: "input:clear", mode: keybindings.ModeInsert},
		{combo: "ctrl+a", action: "cursor:start", mode: keybindings.ModeInsert},
		{combo: "ctrl+e", action: "cursor:end", mode: keybindings.ModeInsert},
	}

	for _, check := range required {
		binding, ok := set.Lookup(check.mode, check.combo)
		if !ok {
			t.Fatalf("expected binding for %s in mode %s", check.combo, check.mode)
		}
		if binding.Action != check.action {
			t.Fatalf("expected %s action for %s, got %s", check.action, check.combo, binding.Action)
		}
	}
}
