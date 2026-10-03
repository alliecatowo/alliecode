package tui

import (
	"strings"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/keybindings"
)

type inputMode string

const (
	inputModeChat          inputMode = "chat"
	inputModeSlash         inputMode = "slash"
	inputModeReference     inputMode = "reference"
	inputModeSearch        inputMode = "search"
	inputModeHistorySearch inputMode = "history-search"
	inputModeQuickOpen     inputMode = "quick-open"
	inputModePermission    inputMode = "permission"
	inputModeModelPicker   inputMode = "model-picker"
	inputModeCommandPanel  inputMode = "command-panel"
)

type inputModeState struct {
	mode    inputMode
	keyMode keybindings.Mode
	hint    string
	context string
}

func (m inputMode) label() string {
	if m == "" {
		return string(inputModeChat)
	}
	return string(m)
}

func (a *App) currentInputMode() inputMode {
	return a.resolveInputModeState().mode
}

func (a *App) resolveInputModeState() inputModeState {
	switch a.activeModalSurface() {
	case modalSurfacePermission:
		return inputModeState{mode: inputModePermission, keyMode: keybindings.ModePermission, hint: "enter confirm  tab/shift+tab cycle  esc deny", context: a.permissionModeContext()}
	case modalSurfaceCommandPanel:
		return inputModeState{mode: inputModeCommandPanel, keyMode: keybindings.ModeCommandPanel, hint: modeHintDrawerApply, context: a.commandPanelModeContext()}
	case modalSurfaceModelPicker:
		return inputModeState{mode: inputModeModelPicker, keyMode: keybindings.ModeModelPicker, hint: modeHintDrawerApply, context: a.modelPickerModeContext()}
	case modalSurfaceSlash:
		return inputModeState{mode: inputModeSlash, keyMode: keybindings.ModeSlash, hint: modeHintDrawerApply, context: a.slashModeContext()}
	case modalSurfaceReference:
		return inputModeState{mode: inputModeReference, keyMode: keybindings.ModeReference, hint: modeHintDrawerInsert, context: a.referenceModeContext()}
	case modalSurfaceSearch:
		switch a.searchModeValue() {
		case searchModeQuickOpen:
			return inputModeState{mode: inputModeQuickOpen, keyMode: keybindings.ModeQuickOpen, hint: modeHintDrawerApply, context: a.searchModeContext()}
		case searchModeHistory:
			return inputModeState{mode: inputModeHistorySearch, keyMode: keybindings.ModeHistory, hint: modeHintDrawerApply, context: a.searchModeContext()}
		default:
			return inputModeState{mode: inputModeSearch, keyMode: keybindings.ModeTimeline, hint: modeHintTimeline, context: a.searchModeContext()}
		}
	default:
		return inputModeState{mode: inputModeChat, keyMode: keybindings.ModeChat, hint: "enter send  shift+tab history  ctrl+j newline", context: a.chatModeContext()}
	}
}

func (a *App) syncInputMode() {
	a.syncModalStack()
	state := a.resolveInputModeState()
	a.setInputModeValue(state.mode)
	a.input.SetHintOverride(state.hint)
}

func (a *App) modeHint() string {
	return a.resolveInputModeState().hint
}

func (a *App) activeContextHint() string {
	return a.resolveInputModeState().context
}

func (a *App) chatModeContext() string {
	selection := commands.RuntimeSelectionTruth(&a.cmdState)
	if !selection.ProviderReady && strings.TrimSpace(selection.ProviderName) != "" {
		return "provider " + selection.ProviderName + " needs login via /login provider " + selection.ProviderName
	}
	if strings.TrimSpace(selection.ProviderName) != "" && strings.TrimSpace(selection.ModelName) == "" {
		return "pick model with /model " + selection.ProviderName + "/<model>"
	}
	if strings.TrimSpace(a.input.Value()) != "" {
		return "input enter send  shift+tab history  ctrl+j newline"
	}
	return "input enter send  shift+tab history  ctrl+j newline"
}

func (a *App) slashModeContext() string {
	if item, ok := a.slashAutocomplete.selectedItem(); ok {
		return "command /" + item.Name + " selected | slash selected /" + item.Name
	}
	if len(a.slashAutocomplete.items) == 1 {
		name := a.slashAutocomplete.items[0].Name
		return "command /" + name + " selected | slash selected /" + name
	}
	return "command palette open"
}

func (a *App) referenceModeContext() string {
	if item, ok := a.selectedReferenceSuggestion(); ok {
		return "reference @" + item.Path
	}
	return "reference palette open"
}

func (a *App) commandPanelModeContext() string {
	if item, ok := a.commandPanel.selectedItem(); ok {
		return "command panel /" + a.commandPanel.panel.Command + " -> " + item.Label
	}
	return "command panel /" + a.commandPanel.panel.Command
}

func (a *App) modelPickerModeContext() string {
	if item, ok := a.selectedModelPickerItem(); ok {
		return "model picker /model " + item.provider + "/" + item.model.Model
	}
	return "model picker open"
}

func (a *App) permissionModeContext() string {
	return a.renderPermissionHistorySummary()
}

func (a *App) searchModeContext() string {
	return "search mode " + a.searchModeValue().label() + " selection=" + a.activeSearchSelectionSummary()
}
