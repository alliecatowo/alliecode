package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/keybindings"
)

func (a *App) resolveKeyAction(msg tea.KeyMsg) string {
	if a.keySet == nil {
		return ""
	}
	mode := a.keybindingMode()
	binding, ok := a.keySet.ResolveActionForInputMode(mode, msg.String())
	if !ok {
		return ""
	}
	return binding.Action
}

func (a *App) handleGlobalSearchAction(msg tea.KeyMsg) bool {
	switch a.resolveKeyAction(msg) {
	case "search:timeline":
		a.startTimelineSearch()
		return true
	case "search:quick_open":
		a.startQuickOpen()
		return true
	case "search:history":
		a.startHistorySearch()
		return true
	}
	return false
}

func (a *App) activeModeStack() []keybindings.Mode {
	return keybindings.InputModeStack(a.keybindingMode())
}
