package tui

import (
	"strings"
)

type timelineRenderModel struct {
	rows        []timelineRenderedRow
	matches     []timelineSearchMatch
	totalLines  int
	contentHint int
}

type timelineRenderedRow struct {
	rowIndex   int
	block      string
	lineOffset int
}

func buildTimelineRenderModel(rows []timelineEntry, query string, width int) timelineRenderModel {
	normQuery := strings.ToLower(strings.TrimSpace(query))
	renderedRows := make([]timelineRenderedRow, 0, len(rows))
	matches := make([]timelineSearchMatch, 0, len(rows))
	lineNo := 0
	contentHint := 0

	for i, row := range rows {
		plain := timelineSearchText(row)
		lowerPlain := ""
		if normQuery != "" {
			lowerPlain = strings.ToLower(plain)
			if !strings.Contains(lowerPlain, normQuery) {
				continue
			}
		}
		block := renderTimelineRow(row, normQuery, width)
		if len(renderedRows) > 0 {
			lineNo += 2
			contentHint += 2
		}
		renderedRows = append(renderedRows, timelineRenderedRow{rowIndex: i, block: block, lineOffset: lineNo})
		rowMatches := timelineSearchMatchesForText(row, plain, lowerPlain, normQuery, lineNo)
		for j := range rowMatches {
			rowMatches[j].rowIndex = i
		}
		matches = append(matches, rowMatches...)
		lineNo += visualLineCount(block, width)
		contentHint += len(block)
	}

	return timelineRenderModel{rows: renderedRows, matches: matches, totalLines: lineNo, contentHint: contentHint}
}

func flattenTimelineRenderModel(model timelineRenderModel, width int) (string, []int, []int, []timelineSearchMatch, int) {
	_ = width
	visible := make([]int, 0, len(model.rows))
	lineOffsets := make([]int, 0, len(model.rows))
	var content strings.Builder
	if model.contentHint > 0 {
		content.Grow(model.contentHint)
	}
	for i, row := range model.rows {
		if i > 0 {
			content.WriteString("\n\n")
		}
		content.WriteString(row.block)
		visible = append(visible, row.rowIndex)
		lineOffsets = append(lineOffsets, row.lineOffset)
	}
	return content.String(), visible, lineOffsets, model.matches, model.totalLines
}
