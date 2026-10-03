package tui

import (
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

func renderNodeTreePlain(root types.RenderNode) string {
	if root.Kind != types.RenderNodeRoot {
		return strings.TrimSpace(renderNodePlain(root))
	}
	blocks := make([]string, 0, len(root.Children))
	for _, child := range root.Children {
		block := strings.TrimSpace(renderNodePlain(child))
		if block != "" {
			blocks = append(blocks, block)
		}
	}
	return strings.Join(blocks, "\n\n")
}

func renderNodePlain(node types.RenderNode) string {
	switch node.Kind {
	case types.RenderNodeSection:
		lines := make([]string, 0, len(node.Children))
		for _, child := range node.Children {
			rendered := strings.TrimSpace(renderNodePlain(child))
			if rendered != "" {
				lines = append(lines, rendered)
			}
		}
		return strings.Join(lines, "\n")
	case types.RenderNodeLine:
		return strings.TrimSpace(node.Text)
	case types.RenderNodeTable:
		rows := make([][]string, 0, len(node.Children))
		for _, child := range node.Children {
			if child.Kind != types.RenderNodeRow {
				continue
			}
			rows = append(rows, append([]string(nil), child.Cells...))
		}
		return strings.Join(renderTableRowsPlain(rows), "\n")
	case types.RenderNodeRoot:
		return renderNodeTreePlain(node)
	default:
		if strings.TrimSpace(node.Text) != "" {
			return strings.TrimSpace(node.Text)
		}
		lines := make([]string, 0, len(node.Children))
		for _, child := range node.Children {
			rendered := strings.TrimSpace(renderNodePlain(child))
			if rendered != "" {
				lines = append(lines, rendered)
			}
		}
		return strings.Join(lines, "\n")
	}
}

func renderTableRowsPlain(rows [][]string) []string {
	if len(rows) == 0 {
		return nil
	}
	columns := len(rows[0])
	if columns == 0 {
		return nil
	}
	widths := make([]int, columns)
	for _, row := range rows {
		for i := 0; i < columns && i < len(row); i++ {
			if len(row[i]) > widths[i] {
				widths[i] = len(row[i])
			}
		}
	}
	out := []string{padTableCells(rows[0], widths), padTableRule(widths)}
	for i := 1; i < len(rows); i++ {
		out = append(out, padTableCells(rows[i], widths))
	}
	return out
}

func padTableCells(cells []string, widths []int) string {
	parts := make([]string, len(widths))
	for i := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		parts[i] = cell + strings.Repeat(" ", widths[i]-len(cell))
	}
	return strings.Join(parts, " | ")
}

func padTableRule(widths []int) string {
	parts := make([]string, len(widths))
	for i, width := range widths {
		parts[i] = strings.Repeat("-", width)
	}
	return strings.Join(parts, "-+-")
}
