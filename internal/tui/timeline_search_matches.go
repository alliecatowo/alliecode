package tui

import "strings"

func timelineSearchMatchesForRow(row timelineEntry, normQuery string, lineOffset int) []timelineSearchMatch {
	plain := timelineSearchText(row)
	lowerPlain := strings.ToLower(plain)
	return timelineSearchMatchesForText(row, plain, lowerPlain, normQuery, lineOffset)
}

func timelineSearchMatchesForText(row timelineEntry, plain, lowerPlain, normQuery string, lineOffset int) []timelineSearchMatch {
	if strings.TrimSpace(normQuery) == "" {
		return nil
	}
	ranges := overlapMatchRangesInsensitiveWithLower(plain, lowerPlain, normQuery)
	if len(ranges) == 0 {
		return nil
	}
	matches := make([]timelineSearchMatch, 0, len(ranges))
	for i, matchRange := range ranges {
		matches = append(matches, timelineSearchMatch{
			lineOffset:  lineOffset,
			occurrence:  i,
			rowMatches:  len(ranges),
			excerpt:     buildSearchExcerpt(plain, matchRange[0], matchRange[1], 48),
			matchReason: timelineRowMatchReason(row),
		})
	}
	return matches
}

func timelineRowMatchReason(row timelineEntry) string {
	switch row.kind {
	case timelineUser:
		return "user"
	case timelineAssistant:
		return "assistant"
	case timelineTool:
		return "tool"
	case timelinePermission:
		return "permission"
	case timelineError:
		return "error"
	default:
		return "row"
	}
}

func buildSearchExcerpt(text string, start, end, radius int) string {
	if text == "" {
		return ""
	}
	if start < 0 {
		start = 0
	}
	if end < start {
		end = start
	}
	if radius <= 0 {
		radius = 32
	}
	left := max(0, start-radius)
	right := min(len(text), end+radius)
	excerpt := strings.TrimSpace(strings.Join(strings.Fields(text[left:right]), " "))
	if excerpt == "" {
		return "(blank)"
	}
	if left > 0 {
		excerpt = "... " + excerpt
	}
	if right < len(text) {
		excerpt += " ..."
	}
	return excerpt
}
