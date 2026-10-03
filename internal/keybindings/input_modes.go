package keybindings

// InputModeStack returns the keybinding precedence for a focused input mode.
func InputModeStack(mode Mode) []Mode {
	switch mode {
	case ModeSlash:
		return ModeStack([]Mode{ModeSlash, ModeAutocomplete, ModeChat}, true)
	case ModeReference:
		return ModeStack([]Mode{ModeReference, ModeAutocomplete, ModeChat}, true)
	case ModeQuickOpen:
		return ModeStack([]Mode{ModeQuickOpen, ModeSearch, ModeAutocomplete, ModeChat}, true)
	case ModeHistory:
		return ModeStack([]Mode{ModeHistory, ModeSearch, ModeAutocomplete, ModeChat}, true)
	case ModeTimeline:
		return ModeStack([]Mode{ModeTimeline, ModeSearch, ModeAutocomplete, ModeChat}, true)
	case ModeModelPicker:
		return ModeStack([]Mode{ModeModelPicker, ModeAutocomplete, ModeSlash, ModeChat}, true)
	case ModeCommandPanel:
		return ModeStack([]Mode{ModeCommandPanel, ModeAutocomplete, ModeChat}, true)
	case ModePermission:
		return ModeStack([]Mode{ModePermission}, true)
	default:
		return ModeStack([]Mode{ModeChat}, true)
	}
}

// ResolveActionForInputMode resolves an action using mode stack precedence.
func (s *Set) ResolveActionForInputMode(mode Mode, combo string) (Binding, bool) {
	return s.ResolveActionInStack(InputModeStack(mode), combo)
}
