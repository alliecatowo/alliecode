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

func renderTimeline(rows []timelineEntry, query string, width int) (string, []int, []int, []timelineSearchMatch, int) {
	model := buildTimelineRenderModel(rows, query, width)
	return flattenTimelineRenderModel(model, width)
}

func timelineSearchText(row timelineEntry) string {
	parts := []string{}
	if row.kind == timelineAssistant {
		assistant := renderTimelineAssistantBody(row)
		body := strings.TrimSpace(assistant.body)
		if body != "" {
			parts = append(parts, body)
		}
		if assistant.usedRawText {
			raw := strings.TrimSpace(row.text)
			if raw != "" && raw != body {
				parts = append(parts, raw)
			}
		}
	} else if text := strings.TrimSpace(row.text); text != "" {
		parts = append(parts, text)
	}
	parts = append(parts, row.toolName, row.toolUseID, row.toolSummary, row.toolInputPreview, string(row.toolState), string(row.permissionState))
	return strings.Join(parts, " ")
}

func renderTimelineRow(row timelineEntry, normQuery string, width int) string {
	if width <= 0 {
		width = 80
	}
	bodyWidth := width - 4
	if bodyWidth < 1 {
		bodyWidth = 1
	}
	indent := "  "
	prefix := fmt.Sprintf("[%02d]", row.turn)
	if row.turn <= 0 {
		prefix = "[..]"
	}

	switch row.kind {
	case timelineUser:
		return userLabelStyle.Render(prefix+" USER") + "\n" + userTextStyle.Render(highlightMatches(row.text, normQuery))
	case timelineAssistant:
		assistant := renderTimelineAssistantBody(row)
		return assistantLabelStyle.Render(prefix+" AI") + "\n" + assistantTextStyle.Render(highlightMatches(assistant.body, normQuery))
	case timelineTool:
		return renderTimelineToolRow(row, prefix, normQuery, bodyWidth, indent)
	case timelinePermission:
		return renderTimelinePermissionRow(row, prefix, normQuery, bodyWidth, indent)
	case timelineError:
		return errorStyle.Render(highlightMatches(prefix+" ERROR "+row.text, normQuery))
	default:
		return highlightMatches(row.text, normQuery)
	}
}

func renderTimelineToolRow(row timelineEntry, prefix, normQuery string, bodyWidth int, indent string) string {
	status := strings.ToUpper(string(row.toolState))
	name := strings.TrimSpace(row.toolName)
	if name == "" {
		name = "tool"
	}
	title := fmt.Sprintf("%s TOOL %s [%s]", prefix, name, status)
	lines := []string{title}
	if row.toolInputBytes > 0 {
		meta := fmt.Sprintf("%soutput: %d chars", indent, row.toolInputBytes)
		if row.toolState == toolProgressRunning {
			meta += " (streaming)"
		}
		lines = append(lines, meta)
	}

	summary := strings.TrimSpace(row.toolSummary)
	preview := strings.TrimSpace(row.toolInputPreview)
	if summary != "" {
		lines = append(lines, indent+"input: "+summary)
	} else if preview != "" {
		lines = append(lines, indent+"input: "+preview)
	}
	if row.toolInputBytes > 0 && preview != "" {
		previewLen := len(strings.TrimSpace(preview))
		if row.toolInputBytes > previewLen {
			lines = append(lines, fmt.Sprintf("%ssummary: %d chars hidden", indent, row.toolInputBytes-previewLen))
		}
	}
	if preview != "" && summary != "" && preview != summary {
		lines = append(lines, indent+"latest: "+preview)
	}
	wrapped := wrapAndClampDisplayLines(lines, bodyWidth, 7, indent+"...")
	for i := range wrapped {
		wrapped[i] = highlightMatches(wrapped[i], normQuery)
	}
	return toolStyle.Render(strings.Join(wrapped, "\n"))
}

func renderTimelinePermissionRow(row timelineEntry, prefix, normQuery string, bodyWidth int, indent string) string {
	status := strings.ToUpper(string(row.permissionState))
	name := strings.TrimSpace(row.toolName)
	if name == "" {
		name = "tool"
	}
	lines := []string{fmt.Sprintf("%s PERM %s [%s]", prefix, name, status)}
	if strings.TrimSpace(row.toolUseID) != "" {
		lines = append(lines, indent+"request: "+row.toolUseID)
	}
	wrapped := wrapAndClampDisplayLines(lines, bodyWidth, 4, indent+"...")
	for i := range wrapped {
		wrapped[i] = highlightMatches(wrapped[i], normQuery)
	}
	return permissionStyle.Render(strings.Join(wrapped, "\n"))
}
