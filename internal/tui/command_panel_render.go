package tui

import (
	"fmt"
	"strings"
)

func (a *App) renderCommandPanel() string {
	if !a.commandPanel.active {
		return ""
	}
	rowWidth := a.width - 2
	if rowWidth < 1 {
		rowWidth = 1
	}
	lines := []string{"drawer: " + a.commandPanel.panel.Title}
	if subtitle := strings.TrimSpace(a.commandPanel.panel.Subtitle); subtitle != "" {
		lines = append(lines, truncateDisplayWidth(subtitle, rowWidth, "..."))
	}
	lines = append(lines, panelDivider(rowWidth))
	if header := strings.TrimSpace(renderIntentsPlain(a.commandPanel.panel.HeaderIntents)); header != "" {
		for _, line := range strings.Split(header, "\n") {
			lines = append(lines, truncateDisplayWidth(line, rowWidth, "..."))
		}
	}
	window := 6
	start := a.commandPanel.offset
	if start < 0 || start >= len(a.commandPanel.panel.Items) {
		start = 0
	}
	end := start + window
	if end > len(a.commandPanel.panel.Items) {
		end = len(a.commandPanel.panel.Items)
	}
	lastSection := ""
	for i := start; i < end; i++ {
		item := a.commandPanel.panel.Items[i]
		section := strings.TrimSpace(item.Section)
		if section != "" && !strings.EqualFold(section, lastSection) {
			lines = append(lines, "")
			lines = append(lines, truncateDisplayWidth("## "+section, rowWidth, "..."))
			lastSection = section
		}
		prefix := "  "
		if i == a.commandPanel.selected {
			prefix = "> "
		}
		meta := strings.TrimSpace(item.Status)
		if meta == "" {
			meta = "item"
		}
		row := fmt.Sprintf("%s%-24s  %-10s  %s", prefix, truncateDisplayWidth(item.Label, 24, "..."), "["+meta+"]", strings.TrimSpace(item.Detail))
		lines = append(lines, truncateDisplayWidth(row, rowWidth, "..."))
	}
	if selected, ok := a.commandPanel.selectedItem(); ok {
		lines = append(lines, "")
		lines = append(lines, truncateDisplayWidth("preview: "+selected.Label, rowWidth, "..."))
		lines = append(lines, truncateDisplayWidth(fmt.Sprintf("selection: %d/%d", a.commandPanel.selected+1, len(a.commandPanel.panel.Items)), rowWidth, "..."))
		if apply := strings.TrimSpace(selected.ApplyInput); apply != "" {
			lines = append(lines, truncateDisplayWidth("apply: "+apply, rowWidth, "..."))
		}
		if preview := strings.TrimSpace(renderIntentsPlain(selected.PreviewIntents)); preview != "" {
			for _, line := range strings.Split(preview, "\n") {
				lines = append(lines, truncateDisplayWidth("  "+line, rowWidth, "..."))
			}
		} else {
			raw := renderTimelineAssistantBody(timelineEntry{text: strings.Join(selected.Preview, "\n")})
			if fallback := strings.TrimSpace(raw.body); fallback != "" {
				for _, line := range strings.Split(fallback, "\n") {
					lines = append(lines, truncateDisplayWidth("  "+line, rowWidth, "..."))
				}
			}
		}
	}
	if len(a.commandPanel.panel.Items) > window {
		remaining := len(a.commandPanel.panel.Items) - end
		if remaining > 0 {
			lines = append(lines, fmt.Sprintf("  ... +%d more", remaining))
		}
	}
	if footer := strings.TrimSpace(renderIntentsPlain(a.commandPanel.panel.FooterIntents)); footer != "" {
		lines = append(lines, "")
		for _, line := range strings.Split(footer, "\n") {
			lines = append(lines, truncateDisplayWidth(line, rowWidth, "..."))
		}
	}
	lines = append(lines, "")
	lines = append(lines, truncateDisplayWidth(drawerNavHelpCommandPanel, rowWidth, "..."))
	return slashAutocompleteStyle.Render(strings.Join(lines, "\n"))
}
