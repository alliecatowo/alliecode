package tui

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "")
}

func runeSafeBackspace(s string) string {
	if s == "" {
		return ""
	}
	_, size := utf8.DecodeLastRuneInString(s)
	if size <= 0 || size > len(s) {
		return ""
	}
	return s[:len(s)-size]
}

func truncateDisplayWidth(s string, maxWidth int, ellipsis string) string {
	if maxWidth <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= maxWidth {
		return s
	}
	ellipsisWidth := runewidth.StringWidth(ellipsis)
	if ellipsisWidth >= maxWidth {
		return fitDisplayWidth(s, maxWidth)
	}
	trimmed := fitDisplayWidth(s, maxWidth-ellipsisWidth)
	return trimmed + ellipsis
}

func fitDisplayWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	var b strings.Builder
	used := 0
	for _, r := range s {
		w := runewidth.RuneWidth(r)
		if w < 0 {
			w = 0
		}
		if used+w > maxWidth {
			break
		}
		b.WriteRune(r)
		used += w
	}
	return b.String()
}

func visualLineCount(rendered string, width int) int {
	if rendered == "" {
		return 1
	}
	if width <= 0 {
		width = 1
	}
	count := 0
	for _, line := range strings.Split(rendered, "\n") {
		lineWidth := lipgloss.Width(line)
		if lineWidth <= 0 {
			count++
			continue
		}
		count += (lineWidth + width - 1) / width
	}
	return count
}

func highlightMatches(s, query string) string {
	ranges := overlapMatchRangesInsensitive(s, query)
	if len(ranges) == 0 {
		return s
	}

	var b strings.Builder
	start := 0
	for _, r := range ranges {
		matchStart := r[0]
		matchEnd := r[1]
		if matchStart < start {
			continue
		}
		b.WriteString(s[start:matchStart])
		b.WriteString(searchHighlightStyle.Render(s[matchStart:matchEnd]))
		start = matchEnd
	}
	b.WriteString(s[start:])
	return b.String()
}

func wrapDisplayWidth(s string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{""}
	}
	if s == "" {
		return []string{""}
	}
	rows := make([]string, 0, 4)
	for _, rawLine := range strings.Split(s, "\n") {
		line := rawLine
		if line == "" {
			rows = append(rows, "")
			continue
		}
		for runewidth.StringWidth(line) > maxWidth {
			line = strings.TrimLeft(line, " ")
			cut := fitDisplayWidth(line, maxWidth)
			if cut == "" {
				break
			}
			rows = append(rows, strings.TrimRight(cut, " "))
			line = strings.TrimLeft(line[len(cut):], " ")
		}
		rows = append(rows, strings.TrimRight(line, " "))
	}
	if len(rows) == 0 {
		return []string{""}
	}
	return rows
}

func wrapAndClampDisplayLines(lines []string, width, limit int, overflowLabel string) []string {
	if width <= 0 {
		width = 1
	}
	if limit <= 0 {
		limit = len(lines)
	}
	wrapped := make([]string, 0, len(lines))
	for _, line := range lines {
		wrapped = append(wrapped, wrapDisplayWidth(line, width)...)
	}
	if len(wrapped) <= limit {
		return wrapped
	}
	if strings.TrimSpace(overflowLabel) == "" {
		overflowLabel = "..."
	}
	clamped := append([]string(nil), wrapped[:max(0, limit-1)]...)
	clamped = append(clamped, truncateDisplayWidth(overflowLabel, width, "..."))
	return clamped
}

func panelDivider(width int) string {
	_ = width
	return strings.Repeat("-", 20)
}

func overlapMatchRangesInsensitive(s, query string) [][]int {
	q := strings.TrimSpace(query)
	if q == "" || s == "" {
		return nil
	}

	lowerS := strings.ToLower(s)
	return overlapMatchRangesInsensitiveWithLower(s, lowerS, q)
}

func overlapMatchRangesInsensitiveWithLower(s, lowerS, query string) [][]int {
	q := strings.TrimSpace(query)
	if q == "" || s == "" || lowerS == "" {
		return nil
	}

	lowerQ := strings.ToLower(q)
	if lowerQ == "" || len(lowerQ) > len(lowerS) {
		return nil
	}

	raw := make([][]int, 0, 4)
	for offset := 0; offset <= len(lowerS)-len(lowerQ); {
		idx := strings.Index(lowerS[offset:], lowerQ)
		if idx < 0 {
			break
		}
		start := offset + idx
		end := start + len(lowerQ)
		raw = append(raw, []int{start, end})
		offset = start + 1
	}
	if len(raw) == 0 {
		return nil
	}

	merged := make([][]int, 0, len(raw))
	for _, r := range raw {
		if len(merged) == 0 {
			merged = append(merged, []int{r[0], r[1]})
			continue
		}
		last := merged[len(merged)-1]
		if r[0] <= last[1] {
			if r[1] > last[1] {
				last[1] = r[1]
			}
			continue
		}
		merged = append(merged, []int{r[0], r[1]})
	}

	return merged
}
