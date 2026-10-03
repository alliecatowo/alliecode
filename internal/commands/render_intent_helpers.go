package commands

import (
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

func resultWithIntents(message string, intents ...types.RenderIntent) Result {
	return Result{Handled: true, Message: message, RenderIntents: cloneRenderIntents(intents)}
}

func cloneRenderIntents(intents []types.RenderIntent) []types.RenderIntent {
	if len(intents) == 0 {
		return nil
	}
	out := make([]types.RenderIntent, 0, len(intents))
	for _, intent := range intents {
		if !intent.HasContent() {
			continue
		}
		out = append(out, intent.Clone())
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func summaryCardIntent(title, summary string, fields ...types.RenderField) types.RenderIntent {
	return types.RenderIntent{
		Kind:    types.RenderIntentSummaryCard,
		Title:   strings.TrimSpace(title),
		Summary: strings.TrimSpace(summary),
		Fields:  append([]types.RenderField(nil), fields...),
	}
}

func groupedListIntent(title, summary string, groups ...types.RenderGroup) types.RenderIntent {
	return types.RenderIntent{
		Kind:    types.RenderIntentGroupedList,
		Title:   strings.TrimSpace(title),
		Summary: strings.TrimSpace(summary),
		Groups:  append([]types.RenderGroup(nil), groups...),
	}
}

func checklistIntent(title, summary string, items ...types.RenderChecklistItem) types.RenderIntent {
	return types.RenderIntent{
		Kind:    types.RenderIntentChecklist,
		Title:   strings.TrimSpace(title),
		Summary: strings.TrimSpace(summary),
		Items:   append([]types.RenderChecklistItem(nil), items...),
	}
}

func tableIntent(title, summary string, columns []string, rows ...types.RenderTableRow) types.RenderIntent {
	return types.RenderIntent{
		Kind:    types.RenderIntentTable,
		Title:   strings.TrimSpace(title),
		Summary: strings.TrimSpace(summary),
		Columns: append([]string(nil), columns...),
		Rows:    append([]types.RenderTableRow(nil), rows...),
	}
}

func detailRowsIntent(title, summary string, rows ...types.RenderDetailRow) types.RenderIntent {
	return types.RenderIntent{
		Kind:       types.RenderIntentDetailRows,
		Title:      strings.TrimSpace(title),
		Summary:    strings.TrimSpace(summary),
		DetailRows: append([]types.RenderDetailRow(nil), rows...),
	}
}

func optionListIntent(title, summary string, options ...types.RenderOption) types.RenderIntent {
	return types.RenderIntent{
		Kind:    types.RenderIntentOptionList,
		Title:   strings.TrimSpace(title),
		Summary: strings.TrimSpace(summary),
		Options: append([]types.RenderOption(nil), options...),
	}
}

func actionHintsIntent(title string, hints ...types.RenderActionHint) types.RenderIntent {
	return types.RenderIntent{
		Kind:  types.RenderIntentActionHints,
		Title: strings.TrimSpace(title),
		Hints: append([]types.RenderActionHint(nil), hints...),
	}
}

func actionListIntent(title, summary string, actions ...types.RenderAction) types.RenderIntent {
	return types.RenderIntent{
		Kind:    types.RenderIntentActionList,
		Title:   strings.TrimSpace(title),
		Summary: strings.TrimSpace(summary),
		Actions: append([]types.RenderAction(nil), actions...),
	}
}

func diagnosticsIntent(title, status, summary string, diagnostics []types.RenderDiagnostic, hints []types.RenderActionHint) types.RenderIntent {
	return types.RenderIntent{
		Kind:        types.RenderIntentDiagnostics,
		Title:       strings.TrimSpace(title),
		Status:      strings.TrimSpace(status),
		Summary:     strings.TrimSpace(summary),
		Diagnostics: append([]types.RenderDiagnostic(nil), diagnostics...),
		Hints:       append([]types.RenderActionHint(nil), hints...),
	}
}

func legacyOutputIntents(message string) []types.RenderIntent {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return nil
	}
	lines := strings.Split(trimmed, "\n")
	tag := strings.TrimSpace(lines[0])
	if len(lines) == 1 || strings.TrimSpace(lines[1]) == "" {
		title := "Command output"
		if tag != "" {
			title = humanizeContractLabel(tag)
		}
		return []types.RenderIntent{summaryCardIntent(title, trimmed)}
	}
	rows := make([]types.RenderDetailRow, 0, len(lines)-1)
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			rows = append(rows, detailRow(humanizeContractLabel(parts[0]), strings.TrimSpace(parts[1]), "", ""))
			continue
		}
		rows = append(rows, detailRow("Detail", line, "", ""))
	}
	title := "Command output"
	if tag != "" {
		title = humanizeContractLabel(tag)
	}
	return []types.RenderIntent{detailRowsIntent(title, "Structured command output.", rows...)}
}

func humanizeContractLabel(raw string) string {
	label := strings.TrimSpace(raw)
	if label == "" {
		return ""
	}
	label = strings.ReplaceAll(label, "_", " ")
	label = strings.ReplaceAll(label, "-", " ")
	words := strings.Fields(strings.ToLower(label))
	for i := range words {
		if words[i] == "" {
			continue
		}
		words[i] = strings.ToUpper(words[i][:1]) + words[i][1:]
	}
	return strings.Join(words, " ")
}

func field(label, value string) types.RenderField {
	return types.RenderField{Label: strings.TrimSpace(label), Value: strings.TrimSpace(value)}
}

func group(title, summary string, items ...string) types.RenderGroup {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			out = append(out, strings.TrimSpace(item))
		}
	}
	return types.RenderGroup{Title: strings.TrimSpace(title), Summary: strings.TrimSpace(summary), Items: out}
}

func checklistItem(label string, done bool, detail string) types.RenderChecklistItem {
	return types.RenderChecklistItem{Label: strings.TrimSpace(label), Done: done, Detail: strings.TrimSpace(detail)}
}

func tableRow(cells ...string) types.RenderTableRow {
	row := make([]string, len(cells))
	for i, cell := range cells {
		row[i] = strings.TrimSpace(cell)
	}
	return types.RenderTableRow{Cells: row}
}

func detailRow(label, value, status, detail string) types.RenderDetailRow {
	return types.RenderDetailRow{
		Label:  strings.TrimSpace(label),
		Value:  strings.TrimSpace(value),
		Status: strings.TrimSpace(status),
		Detail: strings.TrimSpace(detail),
	}
}

func option(label, detail, status, hint string, selected bool) types.RenderOption {
	return types.RenderOption{
		Label:    strings.TrimSpace(label),
		Detail:   strings.TrimSpace(detail),
		Status:   strings.TrimSpace(status),
		Hint:     strings.TrimSpace(hint),
		Selected: selected,
	}
}

func hint(label, command string) types.RenderActionHint {
	return types.RenderActionHint{Label: strings.TrimSpace(label), Command: strings.TrimSpace(command)}
}

func action(label, command, detail, status string) types.RenderAction {
	return types.RenderAction{
		Label:   strings.TrimSpace(label),
		Command: strings.TrimSpace(command),
		Detail:  strings.TrimSpace(detail),
		Status:  strings.TrimSpace(status),
	}
}

func diagnostic(label, status, detail string) types.RenderDiagnostic {
	return types.RenderDiagnostic{Label: strings.TrimSpace(label), Status: strings.TrimSpace(status), Detail: strings.TrimSpace(detail)}
}
