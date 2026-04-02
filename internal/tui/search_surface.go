package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/keybindings"
)

func (a *App) renderSearchLine() string {
	label := fmt.Sprintf("search: %s", a.searchQuery)
	if strings.TrimSpace(a.searchQuery) == "" {
		label = "search: (type to filter)"
	}
	meta := fmt.Sprintf("mode:%s  matches:%d  jump:%s  esc:exit", a.searchMode.label(), a.activeSearchMatchCount(), a.searchJumpHint())
	lines := []string{label, meta}
	if hint := a.renderSearchHints(); strings.TrimSpace(hint) != "" {
		lines = append(lines, hint)
	}
	if summary := a.activeSearchDetailsLine(); strings.TrimSpace(summary) != "" {
		lines = append(lines, summary)
	}
	if pane := a.renderSearchHintPane(); strings.TrimSpace(pane) != "" {
		lines = append(lines, pane)
	}
	if detailPane := a.renderSearchDetailsPane(); strings.TrimSpace(detailPane) != "" {
		lines = append(lines, detailPane)
	}
	switch a.searchMode {
	case searchModeQuickOpen:
		lines = append(lines, a.quickOpen.renderLines(a.width-2, 5)...)
	case searchModeHistory:
		lines = append(lines, a.history.renderLines(a.width-2, 5)...)
	}
	return searchStyle.Render(strings.Join(lines, "\n"))
}

func (a *App) renderSearchHints() string {
	if a.keySet == nil {
		return ""
	}
	hints := a.keySet.HintsForMode(keybindings.ModeSearch, 4)
	if len(hints) == 0 {
		return ""
	}
	line := "keys: " + strings.Join(hints, "  |  ")
	return truncateDisplayWidth(line, a.width-2, "...")
}

func (a *App) activeSearchDetailsLine() string {
	width := a.width - 2
	switch a.searchMode {
	case searchModeQuickOpen:
		return a.quickOpen.selectedSummary(width)
	case searchModeHistory:
		return a.history.selectedSummary(width)
	default:
		return ""
	}
}

func (a *App) searchJumpHint() string {
	switch a.searchMode {
	case searchModeQuickOpen, searchModeHistory:
		return "ctrl+n/ctrl+p up/down tab pgup/pgdown"
	default:
		return "ctrl+n/ctrl+p"
	}
}

func (a *App) renderSearchHintPane() string {
	if a.searchMode == searchModeTimeline && strings.TrimSpace(a.searchQuery) == "" {
		return "hint: type to filter timeline rows, then use ctrl+n / ctrl+p to jump matches"
	}
	if a.searchMode == searchModeQuickOpen {
		return "hint: enter to run selected action, ctrl+r pivots to history"
	}
	if a.searchMode == searchModeHistory {
		return "hint: enter stages selected prompt back into input"
	}
	return ""
}

func (a *App) renderSearchDetailsPane() string {
	width := a.width - 2
	if width < 20 {
		width = 20
	}
	switch a.searchMode {
	case searchModeQuickOpen:
		item, ok := a.quickOpen.selectedItem()
		if !ok {
			return ""
		}
		lines := []string{
			truncateDisplayWidth("quick-open preview:", width, "..."),
			truncateDisplayWidth("  action: "+item.value, width, "..."),
			truncateDisplayWidth("  detail: "+item.detail, width, "..."),
		}
		if hint := strings.TrimSpace(item.hint); hint != "" {
			lines = append(lines, truncateDisplayWidth("  shortcut hint: "+hint, width, "..."))
		}
		return strings.Join(lines, "\n")
	case searchModeHistory:
		return strings.Join(a.history.selectedDetailLines(width), "\n")
	default:
		return ""
	}
}

func (a *App) updateSearchInput(msg tea.KeyMsg) {
	key := strings.ToLower(strings.TrimSpace(msg.String()))
	switch key {
	case "ctrl+u":
		if a.searchQuery == "" {
			return
		}
		a.searchQuery = ""
		if a.searchMode == searchModeTimeline {
			a.matchPos = 0
		}
		a.syncSearchHelpers()
		a.refreshViewport()
		return
	case "ctrl+w", "alt+backspace":
		trimmed := strings.TrimRight(a.searchQuery, " ")
		if trimmed == "" {
			if a.searchQuery == "" {
				return
			}
			a.searchQuery = ""
		} else {
			idx := strings.LastIndex(trimmed, " ")
			if idx < 0 {
				a.searchQuery = ""
			} else {
				a.searchQuery = strings.TrimRight(trimmed[:idx], " ")
			}
		}
		if a.searchMode == searchModeTimeline {
			a.matchPos = 0
		}
		a.syncSearchHelpers()
		a.refreshViewport()
		return
	}

	switch msg.Type {
	case tea.KeyEsc:
		a.exitSearch()
	case tea.KeyEnter:
		if a.applySearchSelection() {
			return
		}
		if !a.shouldExitSearchOnEnter() {
			return
		}
		a.exitSearch()
	case tea.KeyBackspace, tea.KeyDelete:
		if len(a.searchQuery) > 0 {
			a.searchQuery = runeSafeBackspace(a.searchQuery)
			if a.searchMode == searchModeTimeline {
				a.matchPos = 0
			}
			a.syncSearchHelpers()
			a.refreshViewport()
		}
	case tea.KeyRunes:
		a.searchQuery += string(msg.Runes)
		if a.searchMode == searchModeTimeline {
			a.matchPos = 0
		}
		a.syncSearchHelpers()
		a.refreshViewport()
	}
}
