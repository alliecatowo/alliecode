package tui

import (
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

func buildIntentRenderTree(intents []types.RenderIntent) types.RenderNode {
	root := types.RenderNode{Kind: types.RenderNodeRoot}
	for _, intent := range intents {
		if !intent.HasContent() {
			continue
		}
		node := buildIntentNode(intent)
		if len(node.Children) == 0 && strings.TrimSpace(node.Text) == "" {
			continue
		}
		root.Children = append(root.Children, node)
	}
	return root
}

func buildIntentNode(intent types.RenderIntent) types.RenderNode {
	section := types.RenderNode{Kind: types.RenderNodeSection}
	appendHeader := func() {
		appendLineNode(&section, intent.Title)
		if strings.TrimSpace(intent.Status) != "" {
			appendLineNode(&section, "Status: "+strings.TrimSpace(intent.Status))
		}
		appendLineNode(&section, intent.Summary)
	}

	switch intent.Kind {
	case types.RenderIntentSummaryCard:
		appendHeader()
		for _, field := range intent.Fields {
			appendLineNode(&section, strings.TrimSpace(field.Label)+": "+strings.TrimSpace(field.Value))
		}
	case types.RenderIntentGroupedList:
		appendHeader()
		for _, group := range intent.Groups {
			label := strings.TrimSpace(group.Title)
			if summary := strings.TrimSpace(group.Summary); summary != "" {
				if label != "" {
					label += ": " + summary
				} else {
					label = summary
				}
			}
			appendLineNode(&section, label)
			for _, item := range group.Items {
				appendLineNode(&section, "- "+strings.TrimSpace(item))
			}
		}
	case types.RenderIntentChecklist:
		appendHeader()
		for _, item := range intent.Items {
			prefix := "[ ]"
			if item.Done {
				prefix = "[x]"
			}
			line := prefix + " " + strings.TrimSpace(item.Label)
			if detail := strings.TrimSpace(item.Detail); detail != "" {
				line += " - " + detail
			}
			appendLineNode(&section, line)
		}
	case types.RenderIntentTable:
		appendHeader()
		if len(intent.Columns) > 0 {
			table := types.RenderNode{Kind: types.RenderNodeTable}
			table.Children = append(table.Children, types.RenderNode{Kind: types.RenderNodeRow, Cells: append([]string(nil), intent.Columns...)})
			for _, row := range intent.Rows {
				table.Children = append(table.Children, types.RenderNode{Kind: types.RenderNodeRow, Cells: append([]string(nil), row.Cells...)})
			}
			section.Children = append(section.Children, table)
		}
	case types.RenderIntentDetailRows:
		appendHeader()
		for _, row := range intent.DetailRows {
			line := strings.TrimSpace(row.Label)
			if value := strings.TrimSpace(row.Value); value != "" {
				line += ": " + value
			}
			if status := strings.TrimSpace(row.Status); status != "" {
				line += " [" + status + "]"
			}
			if detail := strings.TrimSpace(row.Detail); detail != "" {
				line += " - " + detail
			}
			appendLineNode(&section, line)
		}
	case types.RenderIntentOptionList:
		appendHeader()
		for _, option := range intent.Options {
			prefix := "- "
			if option.Selected {
				prefix = "* "
			}
			line := prefix + strings.TrimSpace(option.Label)
			if status := strings.TrimSpace(option.Status); status != "" {
				line += " [" + status + "]"
			}
			if detail := strings.TrimSpace(option.Detail); detail != "" {
				line += ": " + detail
			}
			if hint := strings.TrimSpace(option.Hint); hint != "" {
				line += " -> " + hint
			}
			appendLineNode(&section, line)
		}
	case types.RenderIntentActionHints:
		appendHeader()
		for _, hint := range intent.Hints {
			line := "- "
			if label := strings.TrimSpace(hint.Label); label != "" {
				line += label + ": "
			}
			line += strings.TrimSpace(hint.Command)
			appendLineNode(&section, line)
		}
	case types.RenderIntentActionList:
		appendHeader()
		for _, action := range intent.Actions {
			line := "- " + strings.TrimSpace(action.Label)
			if command := strings.TrimSpace(action.Command); command != "" {
				line += ": " + command
			}
			if status := strings.TrimSpace(action.Status); status != "" {
				line += " [" + status + "]"
			}
			if detail := strings.TrimSpace(action.Detail); detail != "" {
				line += " - " + detail
			}
			appendLineNode(&section, line)
		}
	case types.RenderIntentDiagnostics:
		appendHeader()
		for _, diag := range intent.Diagnostics {
			line := strings.TrimSpace(diag.Label)
			if status := strings.TrimSpace(diag.Status); status != "" {
				line += " [" + status + "]"
			}
			if detail := strings.TrimSpace(diag.Detail); detail != "" {
				line += ": " + detail
			}
			appendLineNode(&section, line)
		}
		for _, hint := range intent.Hints {
			line := "- "
			if label := strings.TrimSpace(hint.Label); label != "" {
				line += label + ": "
			}
			line += strings.TrimSpace(hint.Command)
			appendLineNode(&section, line)
		}
	default:
		appendHeader()
	}

	return section
}

func appendLineNode(section *types.RenderNode, text string) {
	line := strings.TrimSpace(text)
	if line == "" {
		return
	}
	section.Children = append(section.Children, types.RenderNode{Kind: types.RenderNodeLine, Text: line})
}
