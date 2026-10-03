package tui

import "strings"

func (a *App) renderInputModeIndicator() string {
	width := a.width - 2
	if width < 20 {
		width = 20
	}
	state := a.resolveInputModeState()
	context := state.context
	if idx := strings.Index(context, " | "); idx >= 0 {
		context = strings.TrimSpace(context[:idx])
	}
	parts := []string{"input mode: " + state.mode.label()}
	if context != "" {
		parts = append(parts, context)
	}
	if hint := state.hint; hint != "" && state.context != "input "+hint {
		parts = append(parts, hint)
	}
	line := strings.Join(parts, "  |  ")
	return slashAutocompleteStyle.Render(truncateDisplayWidth(line, width, "..."))
}
