package tui

import (
	"fmt"
	"slices"
	"strings"
)

type historySearchEntry struct {
	text string
}

type historySearchMatch struct {
	index int
	score int
}

type historySearchState struct {
	query    string
	entries  []historySearchEntry
	matches  []historySearchMatch
	selected int
	memory   map[string]string
}

func (s historySearchState) selectedSummary(width int) string {
	if len(s.matches) == 0 {
		return "selected: -"
	}
	if s.selected < 0 || s.selected >= len(s.matches) {
		return fmt.Sprintf("selected: -/%d", len(s.matches))
	}
	match := s.matches[s.selected]
	entry := s.entries[match.index]
	label, detail := historyEntryLabelDetail(entry.text)
	line := fmt.Sprintf("selected: %d/%d  score:%d  %s - %s", s.selected+1, len(s.matches), match.score, label, detail)
	if width <= 0 {
		width = len(line)
	}
	return truncateDisplayWidth(line, width, "...")
}

func (s historySearchState) selectedEntry() (historySearchEntry, bool) {
	if s.selected < 0 || s.selected >= len(s.matches) {
		return historySearchEntry{}, false
	}
	idx := s.matches[s.selected].index
	if idx < 0 || idx >= len(s.entries) {
		return historySearchEntry{}, false
	}
	return s.entries[idx], true
}

func newHistorySearchState(entries []historySearchEntry) historySearchState {
	s := historySearchState{entries: append([]historySearchEntry(nil), entries...), selected: -1}
	s.rebuildMatches()
	return s
}

func (s *historySearchState) setQuery(query string) {
	s.rememberSelection()
	s.query = query
	s.rebuildMatches()
}

func (s *historySearchState) moveSelection(dir int) {
	if len(s.matches) == 0 {
		s.selected = -1
		return
	}
	s.selected = nextMatchPos(s.selected, len(s.matches), dir)
	s.rememberSelection()
}

func (s *historySearchState) rebuildMatches() {
	norm := strings.ToLower(strings.TrimSpace(s.query))
	tokens := strings.Fields(norm)
	prevEntryIndex := -1
	if s.selected >= 0 && s.selected < len(s.matches) {
		prevEntryIndex = s.matches[s.selected].index
	}
	s.matches = s.matches[:0]

	for i, entry := range s.entries {
		score, ok := rankHistoryEntry(entry.text, norm, tokens, i, len(s.entries))
		if !ok {
			continue
		}
		s.matches = append(s.matches, historySearchMatch{index: i, score: score})
	}

	slices.SortFunc(s.matches, func(a, b historySearchMatch) int {
		if a.score != b.score {
			if a.score > b.score {
				return -1
			}
			return 1
		}
		if a.index != b.index {
			if a.index > b.index {
				return -1
			}
			return 1
		}
		return 0
	})

	if len(s.matches) == 0 {
		s.selected = -1
		return
	}
	if prevEntryIndex >= 0 {
		for i, match := range s.matches {
			if match.index == prevEntryIndex {
				s.selected = i
				return
			}
		}
	}
	if remembered := s.recallSelection(norm); remembered != "" {
		for i, match := range s.matches {
			entryText := strings.TrimSpace(strings.ToLower(s.entries[match.index].text))
			if entryText == remembered {
				s.selected = i
				return
			}
		}
	}
	if s.selected < 0 || s.selected >= len(s.matches) {
		s.selected = 0
		return
	}
}

func (s *historySearchState) rememberSelection() {
	if len(s.matches) == 0 || s.selected < 0 || s.selected >= len(s.matches) {
		return
	}
	entry := strings.TrimSpace(strings.ToLower(s.entries[s.matches[s.selected].index].text))
	if entry == "" {
		return
	}
	query := strings.TrimSpace(strings.ToLower(s.query))
	if s.memory == nil {
		s.memory = make(map[string]string, 8)
	}
	s.memory[query] = entry
}

func (s historySearchState) recallSelection(query string) string {
	if len(s.memory) == 0 {
		return ""
	}
	return strings.TrimSpace(s.memory[strings.TrimSpace(strings.ToLower(query))])
}

func (s historySearchState) selectedDetailLines(width int) []string {
	if width <= 0 {
		width = 40
	}
	entry, ok := s.selectedEntry()
	if !ok {
		return []string{"history details: none selected"}
	}
	trimmed := strings.TrimSpace(entry.text)
	if trimmed == "" {
		return []string{"history details: (empty prompt)"}
	}
	rawRows := strings.Split(trimmed, "\n")
	rows := make([]string, 0, len(rawRows))
	for _, row := range rawRows {
		line := strings.TrimSpace(row)
		if line == "" {
			continue
		}
		rows = append(rows, line)
	}
	if len(rows) == 0 {
		rows = append(rows, "(blank)")
	}
	out := make([]string, 0, 4)
	out = append(out, truncateDisplayWidth("history details:", width, "..."))
	maxPreview := min(6, len(rows))
	for i := 0; i < maxPreview; i++ {
		out = append(out, truncateDisplayWidth("  "+rows[i], width, "..."))
	}
	if len(rows) > maxPreview {
		out = append(out, truncateDisplayWidth(fmt.Sprintf("  ... +%d more lines", len(rows)-maxPreview), width, "..."))
	}
	return out
}

func (s historySearchState) renderLines(width, limit int) []string {
	if width <= 0 {
		width = 1
	}
	if limit <= 0 {
		limit = 5
	}

	if len(s.matches) == 0 {
		if strings.TrimSpace(s.query) == "" {
			return []string{"  history: no entries"}
		}
		return []string{"  history: no matching prompts"}
	}

	maxItems := min(limit, len(s.matches))
	lines := make([]string, 0, maxItems+1)
	for i := 0; i < maxItems; i++ {
		match := s.matches[i]
		entry := s.entries[match.index]
		label, detail := historyEntryLabelDetail(entry.text)
		status := fmt.Sprintf("score:%d", match.score)
		lines = append(lines, renderSearchListLine(i == s.selected, label, detail, status, width))
	}
	if len(s.matches) > maxItems {
		lines = append(lines, fmt.Sprintf("  ... +%d more", len(s.matches)-maxItems))
	}
	return lines
}

func historyEntryLabelDetail(text string) (string, string) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "(empty prompt)", "history"
	}
	lines := strings.Split(trimmed, "\n")
	label := strings.TrimSpace(lines[0])
	if label == "" {
		label = "(empty prompt)"
	}
	detail := "1 line"
	if len(lines) > 1 {
		detail = fmt.Sprintf("%d lines", len(lines))
	}
	return label, detail
}

func rankHistoryEntry(text, normQuery string, tokens []string, index, total int) (int, bool) {
	rawText := strings.ToLower(strings.TrimSpace(text))
	normText := strings.Join(strings.Fields(rawText), " ")
	if normQuery == "" {
		return 140 + indexRecencyBoost(index, total), true
	}

	score := 0
	matched := false
	firstLine := rawText
	if idx := strings.IndexByte(firstLine, '\n'); idx >= 0 {
		firstLine = strings.TrimSpace(firstLine[:idx])
	}
	firstLine = strings.Join(strings.Fields(firstLine), " ")

	if normText == normQuery {
		score += 1200
		matched = true
	} else if strings.HasPrefix(normText, normQuery) {
		score += 920
		matched = true
	} else if strings.Contains(normText, normQuery) {
		score += 560
		matched = true
	}

	if strings.HasPrefix(firstLine, normQuery) {
		score += 150
		matched = true
	} else if strings.Contains(firstLine, normQuery) {
		score += 80
		matched = true
	}

	tokenHits := 0
	subsequenceHits := 0
	for _, token := range tokens {
		if token == "" {
			continue
		}
		if strings.Contains(normText, token) {
			score += 140
			tokenHits++
			matched = true
		} else if isSubsequence(normText, token) {
			score += 45
			subsequenceHits++
			matched = true
		}
	}
	if len(tokens) > 1 && tokenHits == len(tokens) {
		score += 120
	}
	if subsequenceHits > 0 && tokenHits == 0 {
		score -= 10
	}

	if !matched && isSubsequence(normText, normQuery) {
		score += 210
		matched = true
	}

	if !matched {
		return 0, false
	}

	score += indexRecencyBoost(index, total)
	return score, true
}

func indexRecencyBoost(index, total int) int {
	if total <= 0 {
		return 0
	}
	if index < 0 {
		index = 0
	}
	if index >= total {
		index = total - 1
	}
	return index + 1
}
