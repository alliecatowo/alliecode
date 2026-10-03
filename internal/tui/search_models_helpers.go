package tui

import (
	"fmt"
	"slices"
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
	offset   int
	memory   map[string]string
}

func newQuickOpenState(items []quickOpenItem) quickOpenState {
	s := quickOpenState{items: append([]quickOpenItem(nil), items...), selected: -1, memory: make(map[string]string, 8)}
	s.rebuildVisible("")
	return s
}

func (s *quickOpenState) setQuery(query string) {
	s.rememberSelection()
	prevQuery := strings.TrimSpace(strings.ToLower(s.query))
	s.query = query
	s.rebuildVisible(prevQuery)
}

func (s *quickOpenState) moveSelection(dir int) {
	if len(s.visible) == 0 {
		s.selected = -1
		s.offset = 0
		return
	}
	s.selected = nextMatchPos(s.selected, len(s.visible), dir)
	s.rememberSelection()
	s.ensureVisible(5)
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

func (s *quickOpenState) rebuildVisible(prevQuery string) {
	norm := strings.ToLower(strings.TrimSpace(s.query))
	tokens := strings.Fields(norm)
	prevItemIndex := -1
	if prevQuery != "" && (strings.HasPrefix(norm, prevQuery) || strings.HasPrefix(prevQuery, norm)) && s.selected >= 0 && s.selected < len(s.visible) {
		prevItemIndex = s.visible[s.selected]
	}
	s.visible = s.visible[:0]
	type rankedQuickOpen struct {
		index int
		score int
	}
	ranked := make([]rankedQuickOpen, 0, len(s.items))
	for i, item := range s.items {
		score, ok := rankQuickOpenItem(item, norm, tokens, i, len(s.items))
		if ok {
			ranked = append(ranked, rankedQuickOpen{index: i, score: score})
		}
	}
	slices.SortFunc(ranked, func(a, b rankedQuickOpen) int {
		if a.score != b.score {
			if a.score > b.score {
				return -1
			}
			return 1
		}
		if a.index < b.index {
			return -1
		}
		if a.index > b.index {
			return 1
		}
		return 0
	})
	for _, item := range ranked {
		s.visible = append(s.visible, item.index)
	}
	if len(s.visible) == 0 {
		s.selected = -1
		s.offset = 0
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
				s.ensureVisible(5)
				return
			}
		}
	}
	if s.selected < 0 || s.selected >= len(s.visible) {
		s.selected = 0
	}
	s.ensureVisible(5)
}

func (s *quickOpenState) ensureVisible(window int) {
	if window <= 0 {
		window = 5
	}
	if s.selected < 0 {
		s.offset = 0
		return
	}
	if s.selected < s.offset {
		s.offset = s.selected
		return
	}
	if s.selected >= s.offset+window {
		s.offset = s.selected - window + 1
	}
	maxOffset := len(s.visible) - window
	if maxOffset < 0 {
		maxOffset = 0
	}
	if s.offset > maxOffset {
		s.offset = maxOffset
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
	value := strings.TrimSpace(strings.ToLower(s.items[idx].value))
	s.memory[query] = value
}

func (s quickOpenState) recallSelection(query string) string {
	return recallClosestSelection(s.memory, query)
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

func rankQuickOpenItem(item quickOpenItem, normQuery string, tokens []string, index, total int) (int, bool) {
	if normQuery == "" {
		return 160 + max(0, total-index), true
	}
	fields := []string{
		strings.ToLower(strings.TrimSpace(item.label)),
		strings.ToLower(strings.TrimSpace(item.detail)),
		strings.ToLower(strings.TrimSpace(item.status)),
		strings.ToLower(strings.TrimSpace(item.keywords)),
		strings.ToLower(strings.TrimSpace(item.value)),
	}
	combined := strings.Join(fields, " ")
	score := 0
	matched := false
	if fields[4] == normQuery {
		score += 1400
		matched = true
	} else if fields[0] == normQuery {
		score += 1320
		matched = true
	}
	if strings.HasPrefix(fields[4], normQuery) {
		score += 640
		matched = true
	}
	if strings.HasPrefix(fields[0], normQuery) {
		score += 520
		matched = true
	}
	for _, field := range fields {
		if field == "" {
			continue
		}
		if strings.Contains(field, normQuery) {
			score += 220
			matched = true
		}
	}
	tokenHits := 0
	for _, token := range tokens {
		if token == "" {
			continue
		}
		if strings.Contains(combined, token) {
			score += 80
			tokenHits++
			matched = true
		}
	}
	if len(tokens) > 1 && tokenHits == len(tokens) {
		score += 120
	}
	if !matched && isSubsequence(combined, normQuery) {
		score += 180
		matched = true
	}
	if !matched {
		return 0, false
	}
	return score + max(0, total-index), true
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

func (s quickOpenState) selectedDetailLines(width int) []string {
	if width <= 0 {
		width = 40
	}
	item, ok := s.selectedItem()
	if !ok {
		return []string{"quick-open preview: none selected"}
	}
	lines := []string{
		"quick-open preview:",
		fmt.Sprintf("  selection: %d/%d", s.selected+1, len(s.visible)),
		"  selected: " + strings.TrimSpace(item.label),
		"  action: " + strings.TrimSpace(item.value),
	}
	if detail := strings.TrimSpace(item.detail); detail != "" {
		lines = append(lines, "  detail: "+detail)
	}
	if status := strings.TrimSpace(item.status); status != "" {
		lines = append(lines, "  section: "+quickOpenSectionLabel(status))
	}
	if hint := strings.TrimSpace(item.hint); hint != "" {
		lines = append(lines, "  shortcut hint: "+hint)
	}
	if keywords := strings.TrimSpace(item.keywords); keywords != "" {
		lines = append(lines, "  search: "+keywords)
	}
	return wrapAndClampDisplayLines(lines, width, 7, "  ...")
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
	start := s.offset
	if start < 0 || start >= len(s.visible) {
		start = 0
	}
	end := start + maxItems
	if end > len(s.visible) {
		end = len(s.visible)
	}
	lastSection := ""
	for i := start; i < end; i++ {
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
	if len(s.visible) > end {
		lines = append(lines, fmt.Sprintf("  ... +%d more", len(s.visible)-end))
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

	base := fmt.Sprintf("%s%-24s", prefix, truncateDisplayWidth(label, 24, "..."))
	if detail != "-" {
		base += "  " + truncateDisplayWidth(detail, 34, "...")
	}
	base += "  [" + status + "]"
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
