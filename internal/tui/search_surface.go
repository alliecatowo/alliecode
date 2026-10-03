package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/keybindings"
)

func (a *App) renderSearchLine() string {
	width := a.width - 2
	if width < 20 {
		width = 20
	}
	label := fmt.Sprintf("search: %s", a.searchQueryValue())
	if strings.TrimSpace(a.searchQueryValue()) == "" {
		label = "search: (type to filter)"
	}
	meta := fmt.Sprintf("mode:%s  matches:%d  jump:%s  esc:exit", a.searchModeValue().label(), a.activeSearchMatchCount(), a.searchJumpHint())
	divider := panelDivider(width)
	lines := []string{truncateDisplayWidth(label, width, "..."), divider, truncateDisplayWidth(meta, width, "...")}
	if hint := a.renderSearchHints(); strings.TrimSpace(hint) != "" {
		lines = append(lines, truncateDisplayWidth(hint, width, "..."))
	}
	if summary := a.activeSearchDetailsLine(); strings.TrimSpace(summary) != "" {
		lines = append(lines, truncateDisplayWidth(summary, width, "..."))
	}
	if pane := a.renderSearchHintPane(); strings.TrimSpace(pane) != "" {
		lines = append(lines, truncateDisplayWidth(pane, width, "..."))
	}

	if footer := a.renderSearchFooterHints(); strings.TrimSpace(footer) != "" {
		lines = append(lines, truncateDisplayWidth(footer, width, "..."))
	}
	if detailPane := a.renderSearchDetailsPane(); strings.TrimSpace(detailPane) != "" {
		for _, row := range strings.Split(detailPane, "\n") {
			for _, wrapped := range wrapDisplayWidth(row, width) {
				lines = append(lines, truncateDisplayWidth(wrapped, width, "..."))
			}
		}
	}

	switch a.searchModeValue() {
	case searchModeQuickOpen:
		for _, row := range a.quickOpen.renderLines(width, 5) {
			lines = append(lines, truncateDisplayWidth(row, width, "..."))
		}
	case searchModeHistory:
		for _, row := range a.history.renderLines(width, 5) {
			lines = append(lines, truncateDisplayWidth(row, width, "..."))
		}
	}

	return searchStyle.Render(strings.Join(lines, "\n"))
}

func (a *App) renderSearchHints() string {
	if a.keySet == nil {
		return ""
	}
	hints := a.keySet.HintsForModes(keybindings.InputModeStack(a.keybindingMode()), 4)
	if len(hints) == 0 {
		return ""
	}
	line := "keys: " + strings.Join(hints, "  |  ")
	width := a.width - 2
	if width < 20 {
		width = 20
	}
	return truncateDisplayWidth(line, width, "...")
}

func (a *App) activeSearchDetailsLine() string {
	width := a.width - 2
	switch a.searchModeValue() {
	case searchModeQuickOpen:
		return a.quickOpen.selectedSummary(width)
	case searchModeHistory:
		return a.history.selectedSummary(width)
	default:
		return ""
	}
}

func (a *App) searchJumpHint() string {
	switch a.searchModeValue() {
	case searchModeQuickOpen, searchModeHistory:
		return "ctrl+n/ctrl+p up/down tab pgup/pgdown"
	default:
		return "ctrl+n/ctrl+p up/down tab shift+tab pgup/pgd..."
	}
}

func (a *App) renderSearchFooterHints() string {
	width := a.width - 2
	if width < 20 {
		width = 20
	}
	switch a.searchModeValue() {
	case searchModeQuickOpen:
		if item, ok := a.quickOpen.selectedItem(); ok {
			return truncateDisplayWidth("active: "+item.label+"  |  "+searchNavHelpSelection, width, "...")
		}
		return truncateDisplayWidth(searchNavHelpSelection, width, "...")
	case searchModeHistory:
		if item, ok := a.history.selectedEntry(); ok {
			line := strings.TrimSpace(item.text)
			if idx := strings.IndexByte(line, '\n'); idx >= 0 {
				line = strings.TrimSpace(line[:idx])
			}
			if line == "" {
				line = "(empty prompt)"
			}
			return truncateDisplayWidth("active: "+line+"  |  "+searchNavHelpSelection, width, "...")
		}
		return truncateDisplayWidth(searchNavHelpSelection, width, "...")
	default:
		return truncateDisplayWidth(searchNavHelpTimeline, width, "...")
	}
}

func (a *App) renderSearchHintPane() string {
	if a.searchModeValue() == searchModeTimeline && strings.TrimSpace(a.searchQueryValue()) == "" {
		return "hint: type to filter timeline rows, then use ctrl+n / ctrl+p to jump matches"
	}
	if a.searchModeValue() == searchModeQuickOpen {
		return "hint: enter applies the selected action immediately"
	}
	if a.searchModeValue() == searchModeHistory {
		return "hint: enter applies the selected history prompt into input"
	}
	return ""
}

func (a *App) renderSearchDetailsPane() string {
	width := a.width - 2
	if width < 20 {
		width = 20
	}
	switch a.searchModeValue() {
	case searchModeQuickOpen:
		return strings.Join(a.quickOpen.selectedDetailLines(width), "\n")
	case searchModeHistory:
		return strings.Join(a.history.selectedDetailLines(width), "\n")
	case searchModeTimeline:
		if strings.TrimSpace(a.searchQueryValue()) == "" || len(a.timelineMatches) == 0 || a.searchMatchPosValue() < 0 || a.searchMatchPosValue() >= len(a.timelineMatches) {
			return ""
		}
		match := a.timelineMatches[a.searchMatchPosValue()]
		lines := []string{
			fmt.Sprintf("timeline details: %d/%d", a.searchMatchPosValue()+1, len(a.timelineMatches)),
			fmt.Sprintf("  row: %d  kind: %s  row hits: %d", match.rowIndex+1, match.matchReason, max(1, match.rowMatches)),
			"  excerpt: " + match.excerpt,
		}
		return strings.Join(wrapAndClampDisplayLines(lines, width, 6, "  ..."), "\n")
	default:
		return ""
	}
}

func (a *App) updateSearchInput(msg tea.KeyMsg) {
	key := strings.ToLower(strings.TrimSpace(msg.String()))
	switch key {
	case "ctrl+u":
		if a.searchQueryValue() == "" {
			return
		}
		a.setSearchQueryValue("")
		if a.searchModeValue() == searchModeTimeline {
			a.setSearchMatchPosValue(0)
		}
		a.syncSearchHelpers()
		a.refreshViewport()
		return
	case "ctrl+w", "alt+backspace":
		trimmed := strings.TrimRight(a.searchQueryValue(), " ")
		if trimmed == "" {
			if a.searchQueryValue() == "" {
				return
			}
			a.setSearchQueryValue("")
		} else {
			idx := strings.LastIndex(trimmed, " ")
			if idx < 0 {
				a.setSearchQueryValue("")
			} else {
				a.setSearchQueryValue(strings.TrimRight(trimmed[:idx], " "))
			}
		}
		if a.searchModeValue() == searchModeTimeline {
			a.setSearchMatchPosValue(0)
		}
		a.syncSearchHelpers()
		a.refreshViewport()
		return
	}

	switch msg.Type {
	case tea.KeyEsc:
		a.exitSearch()
	case tea.KeyEnter:
		if applied, cmd := a.applySearchSelection(); applied {
			if cmd != nil {
				msg := cmd()
				if msg != nil {
					_, _ = a.Update(msg)
				}
			}
			return
		}
		if !a.shouldExitSearchOnEnter() {
			return
		}
		a.exitSearch()
	case tea.KeyBackspace, tea.KeyDelete:
		if len(a.searchQueryValue()) > 0 {
			a.setSearchQueryValue(runeSafeBackspace(a.searchQueryValue()))
			if a.searchModeValue() == searchModeTimeline {
				a.setSearchMatchPosValue(0)
			}
			a.syncSearchHelpers()
			a.refreshViewport()
		}
	case tea.KeyRunes:
		a.setSearchQueryValue(a.searchQueryValue() + string(msg.Runes))
		if a.searchModeValue() == searchModeTimeline {
			a.setSearchMatchPosValue(0)
		}
		a.syncSearchHelpers()
		a.refreshViewport()
	}
}
