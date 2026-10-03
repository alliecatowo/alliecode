package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderStatusLineWithRight(left, right string, width int) string {
	return renderStatusLineWithStyle(left, right, width, statusBarStyle)
}

func renderStatusLineWithStyle(left, right string, width int, style lipgloss.Style) string {
	if width <= 0 {
		width = 1
	}
	contentWidth := width - 2
	if contentWidth < 1 {
		contentWidth = 1
	}
	rightWidth := lipgloss.Width(right)
	leftWidth := lipgloss.Width(left)
	line := ""
	if leftWidth+1+rightWidth <= contentWidth {
		line = left + strings.Repeat(" ", contentWidth-leftWidth-rightWidth) + right
	} else if rightWidth >= contentWidth {
		line = truncateDisplayWidth(right, contentWidth, "...")
	} else {
		availLeft := contentWidth - rightWidth - 1
		if availLeft < 1 {
			availLeft = 1
		}
		line = truncateDisplayWidth(left, availLeft, "...") + " " + right
	}
	return style.Render(" " + line + " ")
}

func (a *App) primaryStatusParts() []string {
	parts := []string{
		a.statusProviderLabel() + "/" + a.statusModelLabel(),
		permissionModeLabel(a.cmdState.PermissionMode),
		"input " + a.inputModeValue().label(),
	}
	if a.stateValue() != stateIdle {
		parts = append(parts, "state "+a.stateLabel())
	}
	return parts
}

func (a *App) primaryStatusInlineDetails() []string {
	parts := []string{"provider:" + a.statusProviderLabel()}
	if mode := permissionModeLabel(a.cmdState.PermissionMode); mode != "default" {
		parts = append(parts, "mode:"+mode)
	}
	if a.turns > 0 {
		parts = append(parts, fmt.Sprintf("turns:%d", a.turns))
	}
	if a.stateValue() == statePermissionPrompt && a.permission.queueTotal > 0 {
		parts = append(parts, fmt.Sprintf("queue:%d/%d", a.permission.queueIndex, a.permission.queueTotal))
	}
	if a.stateValue() == stateSearch {
		parts = append(parts, fmt.Sprintf("matches:%d", a.activeSearchMatchCount()))
	}
	return parts
}

func (a *App) renderSecondaryStatusLine() string {
	parts := a.secondaryStatusParts()
	if len(parts) == 0 {
		return ""
	}
	line := strings.Join(parts, "  |  ")
	width := a.width - 2
	if width < 1 {
		width = 1
	}
	return statusSubtleStyle.Render(" " + truncateDisplayWidth(line, width, "...") + " ")
}

func (a *App) secondaryStatusParts() []string {
	parts := make([]string, 0, 5)
	parts = append(parts, "provider:"+a.statusProviderLabel())
	if permissionModeLabel(a.cmdState.PermissionMode) != "default" {
		parts = append(parts, "mode:"+permissionModeLabel(a.cmdState.PermissionMode))
	}
	if a.inputModeValue() != inputModeChat {
		parts = append(parts, "input:"+a.inputModeValue().label())
	}
	if a.turns > 0 {
		parts = append(parts, fmt.Sprintf("turns:%d", a.turns))
	}
	if a.stateValue() == statePermissionPrompt && a.permission.queueTotal > 0 {
		parts = append(parts, fmt.Sprintf("queue:%d/%d", a.permission.queueIndex, a.permission.queueTotal))
	}
	if a.stateValue() == stateSearch {
		parts = append(parts, fmt.Sprintf("search:%s", a.searchModeValue().label()))
		parts = append(parts, fmt.Sprintf("matches:%d", a.activeSearchMatchCount()))
	}
	if a.commandPanel.active {
		parts = append(parts, "panel:"+strings.TrimSpace(a.commandPanel.panel.Command))
	}
	if a.refAuto.active {
		parts = append(parts, fmt.Sprintf("refs:%d", len(a.refAuto.suggestions)))
	}
	if a.slashAutocomplete.isVisible() {
		parts = append(parts, fmt.Sprintf("slash:%d", len(a.slashAutocomplete.items)))
	}
	return parts
}
