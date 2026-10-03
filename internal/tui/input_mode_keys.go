package tui

import tea "github.com/charmbracelet/bubbletea"

func (a *App) handleModeKey(msg tea.KeyMsg) (handled bool, cmd tea.Cmd) {
	if a.handleGlobalSearchAction(msg) {
		return true, nil
	}

	switch a.activeModalSurface() {
	case modalSurfacePermission:
		var next PermissionModel
		next, cmd = a.permission.Update(msg)
		a.permission = next
		return true, cmd
	case modalSurfaceReference:
		return a.handleReferenceModeKey(msg)
	case modalSurfaceCommandPanel:
		return a.handleCommandPanelModeKey(msg)
	case modalSurfaceModelPicker:
		return a.handleModelPickerModeKey(msg)
	case modalSurfaceSlash:
		return a.handleSlashModeKey(msg)
	case modalSurfaceSearch:
		return a.handleSearchModeKey(msg)
	default:
		return false, nil
	}
}

func (a *App) handleCommandPanelModeKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch {
	case keyMatches(msg, "ctrl+n") || keyMatches(msg, "down") || keyMatches(msg, "tab"):
		a.commandPanel.moveSelection(1)
		return true, nil
	case keyMatches(msg, "ctrl+p") || keyMatches(msg, "up") || keyMatches(msg, "shift+tab"):
		a.commandPanel.moveSelection(-1)
		return true, nil
	case keyMatches(msg, "pgdown") || keyMatches(msg, "pagedown"):
		a.commandPanel.pageSelection(1)
		return true, nil
	case keyMatches(msg, "pgup") || keyMatches(msg, "pageup"):
		a.commandPanel.pageSelection(-1)
		return true, nil
	case keyMatches(msg, "home"):
		a.commandPanel.jumpSelection(false)
		return true, nil
	case keyMatches(msg, "end"):
		a.commandPanel.jumpSelection(true)
		return true, nil
	case keyMatches(msg, "esc") || keyMatches(msg, "left"):
		a.dismissInteractiveCommandPanel()
		a.syncInputMode()
		return true, nil
	case keyMatches(msg, "enter") || keyMatches(msg, "right"):
		if cmd, ok := a.applyInteractiveCommandPanelSelection(); ok {
			a.syncInputMode()
			if cmd == nil {
				return true, nil
			}
			event := cmd()
			if event == nil {
				return true, nil
			}
			_, follow := a.Update(event)
			return true, follow
		}
		return true, nil
	default:
		return false, nil
	}
}

func (a *App) handleReferenceModeKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch {
	case keyMatches(msg, "ctrl+n") || keyMatches(msg, "down") || keyMatches(msg, "tab"):
		a.advanceReferenceSelection(1)
		return true, nil
	case keyMatches(msg, "ctrl+p") || keyMatches(msg, "up") || keyMatches(msg, "shift+tab"):
		a.advanceReferenceSelection(-1)
		return true, nil
	case keyMatches(msg, "pgdown") || keyMatches(msg, "pagedown"):
		a.pageReferenceSelection(1)
		return true, nil
	case keyMatches(msg, "pgup") || keyMatches(msg, "pageup"):
		a.pageReferenceSelection(-1)
		return true, nil
	case keyMatches(msg, "home"):
		a.jumpReferenceSelection(false)
		return true, nil
	case keyMatches(msg, "end"):
		a.jumpReferenceSelection(true)
		return true, nil
	case keyMatches(msg, "enter") || keyMatches(msg, "right"):
		if a.applyReferenceSelection() {
			a.syncInputMode()
		}
		return true, nil
	case keyMatches(msg, "esc") || keyMatches(msg, "left"):
		a.clearReferenceAutocomplete()
		a.syncInputMode()
		return true, nil
	default:
		return false, nil
	}
}

func (a *App) handleModelPickerModeKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch {
	case keyMatches(msg, "ctrl+n") || keyMatches(msg, "down") || keyMatches(msg, "tab"):
		a.moveModelPickerSelection(1)
		return true, nil
	case keyMatches(msg, "ctrl+p") || keyMatches(msg, "up") || keyMatches(msg, "shift+tab"):
		a.moveModelPickerSelection(-1)
		return true, nil
	case keyMatches(msg, "pgdown") || keyMatches(msg, "pagedown"):
		a.pageModelPickerSelection(1)
		return true, nil
	case keyMatches(msg, "pgup") || keyMatches(msg, "pageup"):
		a.pageModelPickerSelection(-1)
		return true, nil
	case keyMatches(msg, "home"):
		a.jumpModelPickerSelection(false)
		return true, nil
	case keyMatches(msg, "end"):
		a.jumpModelPickerSelection(true)
		return true, nil
	case keyMatches(msg, "esc") || keyMatches(msg, "left"):
		a.dismissModelPicker()
		a.input.Focus()
		a.syncInputMode()
		return true, nil
	case keyMatches(msg, "enter") || keyMatches(msg, "right"):
		if submitText, ok := a.applyModelPickerSelection(); ok {
			a.syncInputMode()
			_, follow := a.Update(submitMsg{text: submitText})
			return true, follow
		}
		return true, nil
	default:
		return false, nil
	}
}

func (a *App) handleSlashModeKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch {
	case keyMatches(msg, "ctrl+n") || keyMatches(msg, "down") || keyMatches(msg, "tab"):
		a.slashAutocomplete.moveSelection(1)
		return true, nil
	case keyMatches(msg, "ctrl+p") || keyMatches(msg, "up") || keyMatches(msg, "shift+tab"):
		a.slashAutocomplete.moveSelection(-1)
		return true, nil
	case keyMatches(msg, "pgdown") || keyMatches(msg, "pagedown"):
		a.slashAutocomplete.pageSelection(1)
		return true, nil
	case keyMatches(msg, "pgup") || keyMatches(msg, "pageup"):
		a.slashAutocomplete.pageSelection(-1)
		return true, nil
	case keyMatches(msg, "home"):
		a.slashAutocomplete.jumpSelection(false)
		return true, nil
	case keyMatches(msg, "end"):
		a.slashAutocomplete.jumpSelection(true)
		return true, nil
	case keyMatches(msg, "esc") || keyMatches(msg, "left"):
		a.clearModelPicker()
		a.slashAutocomplete.clear()
		a.syncInputMode()
		return true, nil
	case keyMatches(msg, "enter") || keyMatches(msg, "right"):
		if applied, immediate, submitText := a.applySlashAutocompleteSelection(); applied {
			a.syncInputMode()
			if immediate {
				_, follow := a.Update(submitMsg{text: submitText})
				return true, follow
			}
		}
		return true, nil
	default:
		return false, nil
	}
}

func (a *App) handleSearchModeKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch {
	case keyMatches(msg, "left") || keyMatches(msg, "esc"):
		a.exitSearch()
		return true, nil
	case keyMatches(msg, "right"):
		if applied, cmd := a.applySearchSelection(); applied {
			if cmd != nil {
				event := cmd()
				if event != nil {
					_, follow := a.Update(event)
					return true, follow
				}
			}
			return true, nil
		}
		if a.timelineSearchActive() {
			a.exitSearch()
		}
		return true, nil
	case keyMatches(msg, "ctrl+n") || keyMatches(msg, "down") || keyMatches(msg, "tab"):
		a.advanceSearchSelection(1)
		return true, nil
	case keyMatches(msg, "ctrl+p") || keyMatches(msg, "up") || keyMatches(msg, "shift+tab"):
		a.advanceSearchSelection(-1)
		return true, nil
	case keyMatches(msg, "pgdown") || keyMatches(msg, "pagedown"):
		a.pageSearchSelection(1)
		return true, nil
	case keyMatches(msg, "pgup") || keyMatches(msg, "pageup"):
		a.pageSearchSelection(-1)
		return true, nil
	case keyMatches(msg, "home"):
		a.jumpSearchSelection(false)
		return true, nil
	case keyMatches(msg, "end"):
		a.jumpSearchSelection(true)
		return true, nil
	case keyMatches(msg, "enter"):
		if applied, cmd := a.applySearchSelection(); applied {
			if cmd != nil {
				event := cmd()
				if event != nil {
					_, follow := a.Update(event)
					return true, follow
				}
			}
			return true, nil
		}
		a.exitSearch()
		return true, nil
	default:
		a.updateSearchInput(msg)
		a.syncInputMode()
		return true, nil
	}
}
