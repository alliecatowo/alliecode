package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func (a *App) openInteractiveCommandPanel(name string) bool {
	cmd, ok := a.commands.Lookup(name)
	if !ok {
		return false
	}
	canonical := strings.TrimSpace(cmd.Name())
	panel, ok := commands.InteractivePanelForCommand(canonical, commands.Context{State: &a.cmdState})
	if !ok || len(panel.Items) == 0 {
		return false
	}
	a.commandPanel.activate(panel, "/"+canonical)
	if a.stateValue() == stateSearch {
		a.setState(stateIdle)
	}
	a.input.SetValue("/" + canonical)
	a.clearModelPicker()
	a.clearReferenceAutocomplete()
	a.slashAutocomplete.clear()
	a.recalcLayout()
	return true
}

func (a *App) closeInteractiveCommandPanel() {
	if !a.commandPanel.active {
		return
	}
	a.commandPanel.clear()
	a.recalcLayout()
}

func (a *App) dismissInteractiveCommandPanel() {
	if !a.commandPanel.active {
		return
	}
	previous := a.commandPanel.previousInput
	a.closeInteractiveCommandPanel()
	if strings.TrimSpace(previous) != "" {
		a.input.SetValue(previous)
	}
}

func (a *App) applyInteractiveCommandPanelSelection() (tea.Cmd, bool) {
	item, ok := a.commandPanel.selectedItem()
	if !ok || strings.TrimSpace(item.ApplyInput) == "" {
		return nil, false
	}
	a.closeInteractiveCommandPanel()
	if item.ApplyMode == commands.PanelApplyStage {
		a.input.SetValue(item.ApplyInput)
		return nil, true
	}
	a.input.Reset()
	return func() tea.Msg { return submitMsg{text: item.ApplyInput} }, true
}

func (a *App) applyInteractiveCommandPanelDefault(name string) (applied bool, immediate bool, submitText string) {
	cmd, ok := a.commands.Lookup(name)
	if !ok {
		return false, false, ""
	}
	canonical := strings.TrimSpace(cmd.Name())
	panel, ok := commands.InteractivePanelForCommand(canonical, commands.Context{State: &a.cmdState})
	if !ok || len(panel.Items) == 0 {
		return false, false, ""
	}
	previewState := a.commandPanel
	previewState.activate(panel, "/"+canonical)
	item, ok := previewState.selectedItem()
	if !ok || strings.TrimSpace(item.ApplyInput) == "" {
		return false, false, ""
	}
	if item.ApplyMode == commands.PanelApplyStage {
		a.input.SetValue(item.ApplyInput)
		return true, false, ""
	}
	a.input.Reset()
	return true, true, item.ApplyInput
}

func (a *App) applyInteractiveCommandPanelSlashDefault(name string) (applied bool, immediate bool, submitText string) {
	cmd, ok := a.commands.Lookup(name)
	if !ok {
		return false, false, ""
	}
	canonical := strings.TrimSpace(cmd.Name())
	panel, ok := commands.InteractivePanelForCommand(canonical, commands.Context{State: &a.cmdState})
	if !ok || len(panel.Items) == 0 {
		return false, false, ""
	}
	item := panel.Items[0]
	want := "/" + canonical
	if !strings.EqualFold(strings.TrimSpace(item.ApplyInput), want) {
		return false, false, ""
	}
	if item.ApplyMode == commands.PanelApplyStage {
		a.input.SetValue(item.ApplyInput)
		return true, false, ""
	}
	a.input.Reset()
	return true, true, item.ApplyInput
}
