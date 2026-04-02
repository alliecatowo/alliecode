package tui

import "github.com/alliecatowo/alliecode/internal/keybindings"

func defaultKeybindingSet() *keybindings.Set {
	return keybindings.NewSet([]keybindings.Binding{
		{Combo: "ctrl+f", Action: "search:timeline", Modes: []keybindings.Mode{keybindings.ModeGlobal, keybindings.ModeChat, keybindings.ModeSearch}},
		{Combo: "ctrl+o", Action: "search:quick_open", Modes: []keybindings.Mode{keybindings.ModeGlobal, keybindings.ModeChat, keybindings.ModeSearch}},
		{Combo: "ctrl+r", Action: "search:history", Modes: []keybindings.Mode{keybindings.ModeGlobal, keybindings.ModeChat, keybindings.ModeSearch}},
		{Combo: "ctrl+k ctrl+f", Action: "search:timeline", Modes: []keybindings.Mode{keybindings.ModeGlobal, keybindings.ModeChat, keybindings.ModeSearch}},
		{Combo: "ctrl+k ctrl+o", Action: "search:quick_open", Modes: []keybindings.Mode{keybindings.ModeGlobal, keybindings.ModeChat, keybindings.ModeSearch}},
		{Combo: "ctrl+k ctrl+r", Action: "search:history", Modes: []keybindings.Mode{keybindings.ModeGlobal, keybindings.ModeChat, keybindings.ModeSearch}},
		{Combo: "ctrl+n", Action: "select:next", Modes: []keybindings.Mode{keybindings.ModeSearch, keybindings.ModeAutocomplete}},
		{Combo: "ctrl+p", Action: "select:previous", Modes: []keybindings.Mode{keybindings.ModeSearch, keybindings.ModeAutocomplete}},
		{Combo: "ctrl+n", Action: "history:next", Modes: []keybindings.Mode{keybindings.ModeChat, keybindings.ModeInsert}},
		{Combo: "ctrl+p", Action: "history:previous", Modes: []keybindings.Mode{keybindings.ModeChat, keybindings.ModeInsert}},
		{Combo: "ctrl+j", Action: "input:newline", Modes: []keybindings.Mode{keybindings.ModeChat, keybindings.ModeInsert}},
		{Combo: "ctrl+w", Action: "input:delete_word", Modes: []keybindings.Mode{keybindings.ModeChat, keybindings.ModeInsert, keybindings.ModeVim}},
		{Combo: "ctrl+u", Action: "input:clear", Modes: []keybindings.Mode{keybindings.ModeChat, keybindings.ModeInsert, keybindings.ModeVim}},
		{Combo: "ctrl+a", Action: "cursor:start", Modes: []keybindings.Mode{keybindings.ModeChat, keybindings.ModeInsert}},
		{Combo: "ctrl+e", Action: "cursor:end", Modes: []keybindings.Mode{keybindings.ModeChat, keybindings.ModeInsert}},
		{Combo: "alt+backspace", Action: "input:delete_word", Modes: []keybindings.Mode{keybindings.ModeChat, keybindings.ModeInsert}},
		{Combo: "ctrl+[", Action: "vim:normal", Modes: []keybindings.Mode{keybindings.ModeVim, keybindings.ModeInsert}},
		{Combo: "ctrl+k ctrl+v", Action: "vim:toggle", Modes: []keybindings.Mode{keybindings.ModeGlobal, keybindings.ModeChat}},
		{Combo: "home", Action: "select:first", Modes: []keybindings.Mode{keybindings.ModeSearch, keybindings.ModeAutocomplete}},
		{Combo: "end", Action: "select:last", Modes: []keybindings.Mode{keybindings.ModeSearch, keybindings.ModeAutocomplete}},
		{Combo: "pgup", Action: "select:page_up", Modes: []keybindings.Mode{keybindings.ModeSearch, keybindings.ModeAutocomplete}},
		{Combo: "pgdown", Action: "select:page_down", Modes: []keybindings.Mode{keybindings.ModeSearch, keybindings.ModeAutocomplete}},
		{Combo: "tab", Action: "select:next", Modes: []keybindings.Mode{keybindings.ModeSearch, keybindings.ModeAutocomplete, keybindings.ModePermission}},
		{Combo: "shift+tab", Action: "select:previous", Modes: []keybindings.Mode{keybindings.ModeSearch, keybindings.ModeAutocomplete, keybindings.ModePermission}},
		{Combo: "enter", Action: "confirm:accept", Modes: []keybindings.Mode{keybindings.ModeSearch, keybindings.ModeAutocomplete, keybindings.ModePermission, keybindings.ModeChat}},
		{Combo: "esc", Action: "dismiss", Modes: []keybindings.Mode{keybindings.ModeSearch, keybindings.ModeAutocomplete, keybindings.ModePermission}},
	})
}
