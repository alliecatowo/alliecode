package tui

import "strings"

type panelSurface struct {
	key      string
	content  string
	minLines int
	maxLines int
}

const (
	searchPanelMinLines      = 4
	searchPanelMaxLines      = 10
	permissionPanelMinLines  = 5
	permissionPanelMaxLines  = 12
	commandPanelMinLines     = 6
	commandPanelMaxLines     = 12
	modelPickerPanelMinLines = 6
	modelPickerPanelMaxLines = 10
	slashPanelMinLines       = 4
	slashPanelMaxLines       = 7
	referencePanelMinLines   = 4
	referencePanelMaxLines   = 8
)

func (a *App) activityPanels() []panelSurface {
	panels := make([]panelSurface, 0, 3)
	switch a.stateValue() {
	case stateThinking:
		panels = append(panels, panelSurface{key: "thinking", content: a.spinner.View(), minLines: 1, maxLines: 2})
	case stateStreaming:
		panels = append(panels, panelSurface{key: "streaming", content: a.renderStreamingLine(), minLines: 1, maxLines: 2})
	default:
		switch a.activeModalSurface() {
		case modalSurfacePermission:
			panels = append(panels, panelSurface{key: "permission", content: a.permission.View(), minLines: permissionPanelMinLines, maxLines: permissionPanelMaxLines})
		case modalSurfaceSearch:
			panels = append(panels, panelSurface{key: "search", content: a.renderSearchLine(), minLines: searchPanelMinLines, maxLines: searchPanelMaxLines})
		}
	}
	return pruneEmptyPanels(panels)
}

func (a *App) drawerPanels() []panelSurface {
	if a.stateValue() != stateIdle && a.stateValue() != stateSearch {
		return nil
	}
	panels := make([]panelSurface, 0, 1)
	switch a.activeModalSurface() {
	case modalSurfaceCommandPanel:
		panels = append(panels, panelSurface{key: "command-panel", content: a.renderCommandPanel(), minLines: commandPanelMinLines, maxLines: commandPanelMaxLines})
	case modalSurfaceModelPicker:
		panels = append(panels, panelSurface{key: "model-picker", content: a.renderModelPicker(), minLines: modelPickerPanelMinLines, maxLines: modelPickerPanelMaxLines})
	case modalSurfaceSlash:
		panels = append(panels, panelSurface{key: "slash", content: a.renderSlashAutocomplete(), minLines: slashPanelMinLines, maxLines: slashPanelMaxLines})
	case modalSurfaceReference:
		panels = append(panels, panelSurface{key: "references", content: a.renderReferenceAutocomplete(), minLines: referencePanelMinLines, maxLines: referencePanelMaxLines})
	}
	return pruneEmptyPanels(panels)
}

func (a *App) composerPanels() []panelSurface {
	if a.stateValue() != stateIdle && a.stateValue() != stateSearch {
		return nil
	}
	panels := make([]panelSurface, 0, 3)
	if hint := strings.TrimSpace(a.renderCommandContextHint()); hint != "" {
		panels = append(panels, panelSurface{key: "input-hint", content: hint, minLines: 1, maxLines: 2})
	}
	panels = append(panels, panelSurface{key: "input-mode", content: a.renderInputModeIndicator(), minLines: 1, maxLines: 1})
	panels = append(panels, panelSurface{key: "input", content: a.input.View(), minLines: 1, maxLines: 4})
	return pruneEmptyPanels(panels)
}

func (a *App) overlayPanels() []panelSurface {
	panels := make([]panelSurface, 0, 2)
	panels = append(panels, a.activityPanels()...)
	panels = append(panels, a.drawerPanels()...)
	return pruneEmptyPanels(panels)
}

func (a *App) inputPanels() []panelSurface {
	panels := make([]panelSurface, 0, 4)
	panels = append(panels, a.drawerPanels()...)
	panels = append(panels, a.composerPanels()...)
	return pruneEmptyPanels(panels)
}

func (a *App) buddyPanels() []panelSurface {
	if a.shouldRenderBuddyFull() {
		content := a.renderBuddySpriteBlock() + "\n" + a.renderBuddyBubbleLine()
		return []panelSurface{{key: "buddy", content: content, minLines: 1, maxLines: 9}}
	}
	if compact := strings.TrimSpace(a.renderBuddyCompactLine()); compact != "" {
		return []panelSurface{{key: "buddy-compact", content: compact, minLines: 1, maxLines: 1}}
	}
	return nil
}

func (a *App) statusPanels() []panelSurface {
	panels := []panelSurface{{key: "status-primary", content: a.renderStatusBar(), minLines: 1, maxLines: 1}}
	if a.showContextualStatusDetails() {
		if a.shouldShowStatusSecondaryLine() {
			if secondary := strings.TrimSpace(a.renderSecondaryStatusLine()); secondary != "" {
				panels = append(panels, panelSurface{key: "status-secondary", content: secondary, minLines: 1, maxLines: 1})
			}
		}
		if a.shouldShowStatusHintsPanel() {
			if hints := strings.TrimSpace(a.renderStatusHints()); hints != "" {
				panels = append(panels, panelSurface{key: "status-hints", content: hints, minLines: 1, maxLines: 2})
			}
		}
		if a.shouldShowStatusRuntimePane() {
			if panes := strings.TrimSpace(a.renderStatusRuntimePanes()); panes != "" {
				panels = append(panels, panelSurface{key: "status-runtime", content: panes, minLines: 1, maxLines: 4})
			}
		}
	}
	return pruneEmptyPanels(panels)
}

func pruneEmptyPanels(panels []panelSurface) []panelSurface {
	filtered := panels[:0]
	for _, panel := range panels {
		if strings.TrimSpace(panel.content) == "" {
			continue
		}
		filtered = append(filtered, panel)
	}
	return filtered
}

func (a *App) panelStackHeight(panels []panelSurface, width int) int {
	if width <= 0 {
		width = 1
	}
	total := 0
	visible := 0
	for _, panel := range panels {
		if strings.TrimSpace(panel.content) == "" {
			continue
		}
		if visible > 0 {
			total++
		}
		lines := visualLineCount(panel.content, width)
		if panel.minLines > 0 && lines < panel.minLines {
			lines = panel.minLines
		}
		if panel.maxLines > 0 && lines > panel.maxLines {
			lines = panel.maxLines
		}
		total += max(0, lines)
		visible++
	}
	return total
}

func (a *App) appendPanels(sections []string, panels []panelSurface) []string {
	added := 0
	for _, panel := range panels {
		if strings.TrimSpace(panel.content) == "" {
			continue
		}
		if added > 0 {
			sections = append(sections, "")
		}
		sections = append(sections, panel.content)
		added++
	}
	return sections
}
