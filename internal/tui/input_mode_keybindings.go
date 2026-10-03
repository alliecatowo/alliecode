package tui

import "github.com/alliecatowo/alliecode/internal/keybindings"

func (a *App) keybindingMode() keybindings.Mode {
	return a.resolveInputModeState().keyMode
}
