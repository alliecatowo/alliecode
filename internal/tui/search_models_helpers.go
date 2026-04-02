package tui

import (
	"fmt"
	"strings"
)

type quickOpenItem struct {
	label    string
	detail   string
	status   string
	keywords string
	value    string
	hint     string
}

type quickOpenState struct {
	query    string
	items    []quickOpenItem
	visible  []int
	selected int
	memory   map[string]string
}

func newQuickOpenState(items []quickOpenItem) quickOpenState {
	s := quickOpenState{items: append([]quickOpenItem(nil), items...), selected: -1}
	s.rebuildVisible()
	return s
}

func (s *quickOpenState) setQuery(query string) {
	s.rememberSelection()
	s.query = query
	s.rebuildVisible()
}

func (s *quickOpenState) moveSelection(dir int) {
	if len(s.visible) == 0 {
		s.selected = -1
		return
	}
	s.selected = nextMatchPos(s.selected, len(s.visible), dir)
	s.rememberSelection()
}

func (s quickOpenState) selectedItem() (quickOpenItem, bool) {
	if s.selected < 0 || s.selected >= len(s.visible) {
		return quickOpenItem{}, false
	}
	idx := s.visible[s.selected]
	if idx < 0 || idx >= len(s.items) {
		return quickOpenItem{}, false
	}
	return s.items[idx], true
}

func (s *quickOpenState) rebuildVisible() {
	norm := strings.ToLower(strings.TrimSpace(s.query))
	prevItemIndex := -1
	if s.selected >= 0 && s.selected < len(s.visible) {
		prevItemIndex = s.visible[s.selected]
	}
	s.visible = s.visible[:0]
	for i, item := range s.items {
		if norm == "" || quickOpenMatches(item, norm) {
			s.visible = append(s.visible, i)
		}
	}
	if len(s.visible) == 0 {
		s.selected = -1
		return
	}
	if prevItemIndex >= 0 {
		for i, idx := range s.visible {
			if idx == prevItemIndex {
				s.selected = i
				return
			}
		}
	}
	if remembered := s.recallSelection(norm); remembered != "" {
		for i, idx := range s.visible {
			if strings.EqualFold(strings.TrimSpace(s.items[idx].value), remembered) {
				s.selected = i
				return
			}
		}
	}
	if s.selected < 0 || s.selected >= len(s.visible) {
		s.selected = 0
	}
}

func (s *quickOpenState) rememberSelection() {
	if len(s.visible) == 0 || s.selected < 0 || s.selected >= len(s.visible) {
		return
	}
	idx := s.visible[s.selected]
	if idx < 0 || idx >= len(s.items) {
		return
	}
	query := strings.TrimSpace(strings.ToLower(s.query))
	if s.memory == nil {
		s.memory = make(map[string]string, 8)
	}
	s.memory[query] = strings.TrimSpace(strings.ToLower(s.items[idx].value))
}

func (s quickOpenState) recallSelection(query string) string {
	if len(s.memory) == 0 {
		return ""
	}
	return strings.TrimSpace(s.memory[strings.TrimSpace(strings.ToLower(query))])
}

func quickOpenMatches(item quickOpenItem, normQuery string) bool {
	fields := []string{item.label, item.detail, item.status, item.keywords, item.value}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(field), normQuery) {
			return true
		}
	}

	compact := strings.ToLower(strings.Join([]string{item.label, item.keywords, item.value}, " "))
	return isSubsequence(compact, normQuery)
}

func quickOpenSectionLabel(status string) string {
	norm := strings.ToLower(strings.TrimSpace(status))
	switch norm {
	case "navigation":
		return "Navigation"
	case "command":
		return "Commands"
	case "history":
		return "History"
	case "workflow":
		return "Workflow"
	case "tools", "tool":
		return "Tools"
	default:
		if norm == "" {
			return "Other"
		}
		return strings.ToUpper(norm[:1]) + norm[1:]
	}
}

func (s quickOpenState) selectedSummary(width int) string {
	if len(s.visible) == 0 {
		return "selected: -"
	}
	if s.selected < 0 || s.selected >= len(s.visible) {
		return fmt.Sprintf("selected: -/%d", len(s.visible))
	}
	item, ok := s.selectedItem()
	if !ok {
		return fmt.Sprintf("selected: %d/%d", s.selected+1, len(s.visible))
	}
	detail := strings.TrimSpace(item.detail)
	if detail == "" {
		detail = item.value
	}
	if strings.TrimSpace(detail) == "" {
		detail = "ready"
	}
	line := fmt.Sprintf("selected: %d/%d  %s - %s", s.selected+1, len(s.visible), item.label, detail)
	if width <= 0 {
		width = len(line)
	}
	return truncateDisplayWidth(line, width, "...")
}

func (s quickOpenState) renderLines(width, limit int) []string {
	if width <= 0 {
		width = 1
	}
	if limit <= 0 {
		limit = 5
	}

	if len(s.visible) == 0 {
		if strings.TrimSpace(s.query) == "" {
			return []string{"  quick-open: no actions available"}
		}
		return []string{"  quick-open: no matching actions"}
	}

	maxItems := min(limit, len(s.visible))
	lines := make([]string, 0, maxItems+4)
	lastSection := ""
	for i := 0; i < maxItems; i++ {
		item := s.items[s.visible[i]]
		section := quickOpenSectionLabel(item.status)
		if section != lastSection {
			lines = append(lines, "  "+section+":")
			lastSection = section
		}
		detail := strings.TrimSpace(item.detail)
		if detail == "" {
			detail = item.value
		}
		status := strings.TrimSpace(item.status)
		if status == "" {
			status = "ready"
		}
		if hint := strings.TrimSpace(item.hint); hint != "" {
			status += ", " + hint
		}
		lines = append(lines, renderSearchListLine(i == s.selected, item.label, detail, status, width))
	}
	if len(s.visible) > maxItems {
		lines = append(lines, fmt.Sprintf("  ... +%d more", len(s.visible)-maxItems))
	}
	return lines
}

func renderSearchListLine(selected bool, label, detail, status string, width int) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	label = strings.TrimSpace(label)
	if label == "" {
		label = "(unnamed)"
	}
	detail = strings.TrimSpace(detail)
	if detail == "" {
		detail = "-"
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = "-"
	}

	base := fmt.Sprintf("%s%s", prefix, label)
	if detail != "-" {
		base += " - " + detail
	}
	base += " [" + status + "]"
	return truncateDisplayWidth(base, width, "...")
}

func isSubsequence(text, query string) bool {
	if query == "" {
		return true
	}
	j := 0
	for i := 0; i < len(text) && j < len(query); i++ {
		if text[i] == query[j] {
			j++
		}
	}
	return j == len(query)
}
