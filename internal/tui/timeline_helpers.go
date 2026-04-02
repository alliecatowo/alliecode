package tui

import (
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/permissions"
)

func buildToolPreview(existing, delta string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	next := strings.TrimSpace(existing + delta)
	return truncateDisplayWidth(next, maxLen, "...")
}

func nextMatchPos(current, total, dir int) int {
	if total <= 0 {
		return 0
	}
	if current < 0 || current >= total {
		if dir < 0 {
			return total - 1
		}
		return 0
	}
	if dir == 0 {
		return current
	}
	step := 1
	if dir < 0 {
		step = -1
	}
	next := current + step
	if next < 0 {
		next = total - 1
	}
	if next >= total {
		next = 0
	}
	return next
}

func permissionModeLabel(mode permissions.Mode) string {
	switch mode {
	case permissions.ModePlan:
		return "plan"
	case permissions.ModeDefault:
		return "default"
	case permissions.ModeAuto:
		return "auto"
	case permissions.ModeBypass:
		return "bypass"
	default:
		return "default"
	}
}

func renderTimeline(rows []timelineEntry, query string, width int) (string, []int, []int, int) {
	normQuery := strings.ToLower(strings.TrimSpace(query))
	visible := make([]int, 0, len(rows))
	lineOffsets := make([]int, 0, len(rows))
	rendered := make([]string, 0, len(rows))
	lineNo := 0

	for i, row := range rows {
		plain := timelineSearchText(row)
		if normQuery != "" && !strings.Contains(strings.ToLower(plain), normQuery) {
			continue
		}
		block := renderTimelineRow(row, normQuery)
		if len(rendered) > 0 {
			lineNo += 2
		}
		lineOffsets = append(lineOffsets, lineNo)
		rendered = append(rendered, block)
		visible = append(visible, i)
		lineNo += visualLineCount(block, width)
	}

	content := strings.Join(rendered, "\n\n")
	return content, visible, lineOffsets, visualLineCount(content, width)
}

func timelineSearchText(row timelineEntry) string {
	parts := []string{row.text, row.toolName, row.toolUseID, row.toolInputPreview, string(row.toolState), string(row.permissionState)}
	return strings.Join(parts, " ")
}

func renderTimelineRow(row timelineEntry, normQuery string) string {
	prefix := fmt.Sprintf("[%02d]", row.turn)
	if row.turn <= 0 {
		prefix = "[..]"
	}

	switch row.kind {
	case timelineUser:
		return userLabelStyle.Render(prefix+" USER") + "\n" + userTextStyle.Render(highlightMatches(row.text, normQuery))
	case timelineAssistant:
		return assistantLabelStyle.Render(prefix+" AI") + "\n" + assistantTextStyle.Render(highlightMatches(row.text, normQuery))
	case timelineTool:
		status := strings.ToUpper(string(row.toolState))
		line := fmt.Sprintf("%s TOOL %s [%s]", prefix, row.toolName, status)
		if strings.TrimSpace(row.toolInputPreview) != "" {
			line += " " + row.toolInputPreview
		}
		return toolStyle.Render(highlightMatches(line, normQuery))
	case timelinePermission:
		status := strings.ToUpper(string(row.permissionState))
		line := fmt.Sprintf("%s PERM %s [%s]", prefix, row.toolName, status)
		return permissionStyle.Render(highlightMatches(line, normQuery))
	case timelineError:
		return errorStyle.Render(highlightMatches(prefix+" ERROR "+row.text, normQuery))
	default:
		return highlightMatches(row.text, normQuery)
	}
}
