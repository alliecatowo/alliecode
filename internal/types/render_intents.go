package types

import "strings"

// RenderIntentKind identifies a structured TUI render target.
type RenderIntentKind string

const (
	RenderIntentSummaryCard RenderIntentKind = "summary_card"
	RenderIntentGroupedList RenderIntentKind = "grouped_list"
	RenderIntentChecklist   RenderIntentKind = "checklist"
	RenderIntentTable       RenderIntentKind = "table"
	RenderIntentDetailRows  RenderIntentKind = "detail_rows"
	RenderIntentOptionList  RenderIntentKind = "option_list"
	RenderIntentActionHints RenderIntentKind = "action_hints"
	RenderIntentActionList  RenderIntentKind = "action_list"
	RenderIntentDiagnostics RenderIntentKind = "diagnostics_panel"
	RenderIntentContract    RenderIntentKind = "contract"
)

// RenderField is a labeled value in a summary card.
type RenderField struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// RenderGroup is a titled collection of related rows.
type RenderGroup struct {
	Title   string   `json:"title,omitempty"`
	Summary string   `json:"summary,omitempty"`
	Items   []string `json:"items,omitempty"`
}

// RenderChecklistItem is a checklist row with optional detail text.
type RenderChecklistItem struct {
	Label  string `json:"label"`
	Done   bool   `json:"done,omitempty"`
	Detail string `json:"detail,omitempty"`
}

// RenderTableRow is a single table row.
type RenderTableRow struct {
	Cells []string `json:"cells,omitempty"`
}

// RenderDetailRow is a labeled row with optional value, state, and supporting text.
type RenderDetailRow struct {
	Label  string `json:"label"`
	Value  string `json:"value,omitempty"`
	Status string `json:"status,omitempty"`
	Detail string `json:"detail,omitempty"`
}

// RenderOption is a selectable or previewable row in a model/provider style list.
type RenderOption struct {
	Label    string `json:"label"`
	Detail   string `json:"detail,omitempty"`
	Status   string `json:"status,omitempty"`
	Hint     string `json:"hint,omitempty"`
	Selected bool   `json:"selected,omitempty"`
}

// RenderActionHint is a suggested follow-up action.
type RenderActionHint struct {
	Label   string `json:"label,omitempty"`
	Command string `json:"command,omitempty"`
}

// RenderAction is a richer suggested follow-up action with detail and state.
type RenderAction struct {
	Label   string `json:"label,omitempty"`
	Command string `json:"command,omitempty"`
	Detail  string `json:"detail,omitempty"`
	Status  string `json:"status,omitempty"`
}

// RenderDiagnostic is a diagnostic row with state and detail.
type RenderDiagnostic struct {
	Label  string `json:"label"`
	Status string `json:"status,omitempty"`
	Detail string `json:"detail,omitempty"`
}

// RenderContractEntry is one key/value line from legacy command contracts.
type RenderContractEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// RenderIntent is a typed rendering request emitted by slash commands.
type RenderIntent struct {
	Kind        RenderIntentKind      `json:"kind"`
	Title       string                `json:"title,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Status      string                `json:"status,omitempty"`
	Fields      []RenderField         `json:"fields,omitempty"`
	Groups      []RenderGroup         `json:"groups,omitempty"`
	Items       []RenderChecklistItem `json:"items,omitempty"`
	Columns     []string              `json:"columns,omitempty"`
	Rows        []RenderTableRow      `json:"rows,omitempty"`
	DetailRows  []RenderDetailRow     `json:"detail_rows,omitempty"`
	Options     []RenderOption        `json:"options,omitempty"`
	Hints       []RenderActionHint    `json:"hints,omitempty"`
	Actions     []RenderAction        `json:"actions,omitempty"`
	Diagnostics []RenderDiagnostic    `json:"diagnostics,omitempty"`
	Contract    []RenderContractEntry `json:"contract,omitempty"`
	ContractTag string                `json:"contract_tag,omitempty"`
}

// HasContent reports whether the intent has renderable content beyond kind.
func (i RenderIntent) HasContent() bool {
	if strings.TrimSpace(i.Title) != "" || strings.TrimSpace(i.Summary) != "" || strings.TrimSpace(i.Status) != "" {
		return true
	}
	return len(i.Fields) > 0 || len(i.Groups) > 0 || len(i.Items) > 0 || len(i.Columns) > 0 || len(i.Rows) > 0 || len(i.DetailRows) > 0 || len(i.Options) > 0 || len(i.Hints) > 0 || len(i.Actions) > 0 || len(i.Diagnostics) > 0 || len(i.Contract) > 0 || strings.TrimSpace(i.ContractTag) != ""
}

// Clone returns a deep-enough copy safe for command/result handoff.
func (i RenderIntent) Clone() RenderIntent {
	clone := i
	clone.Fields = append([]RenderField(nil), i.Fields...)
	clone.Groups = append([]RenderGroup(nil), i.Groups...)
	for gi := range clone.Groups {
		clone.Groups[gi].Items = append([]string(nil), clone.Groups[gi].Items...)
	}
	clone.Items = append([]RenderChecklistItem(nil), i.Items...)
	clone.Columns = append([]string(nil), i.Columns...)
	clone.Rows = append([]RenderTableRow(nil), i.Rows...)
	for ri := range clone.Rows {
		clone.Rows[ri].Cells = append([]string(nil), clone.Rows[ri].Cells...)
	}
	clone.DetailRows = append([]RenderDetailRow(nil), i.DetailRows...)
	clone.Options = append([]RenderOption(nil), i.Options...)
	clone.Hints = append([]RenderActionHint(nil), i.Hints...)
	clone.Actions = append([]RenderAction(nil), i.Actions...)
	clone.Diagnostics = append([]RenderDiagnostic(nil), i.Diagnostics...)
	clone.Contract = append([]RenderContractEntry(nil), i.Contract...)
	return clone
}
