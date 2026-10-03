package types

import "testing"

func TestRenderIntentKindsStable(t *testing.T) {
	if RenderIntentSummaryCard != "summary_card" {
		t.Fatalf("unexpected summary card kind: %q", RenderIntentSummaryCard)
	}
	if RenderIntentDetailRows != "detail_rows" {
		t.Fatalf("unexpected detail rows kind: %q", RenderIntentDetailRows)
	}
	if RenderIntentActionList != "action_list" {
		t.Fatalf("unexpected action list kind: %q", RenderIntentActionList)
	}
	if RenderIntentDiagnostics != "diagnostics_panel" {
		t.Fatalf("unexpected diagnostics kind: %q", RenderIntentDiagnostics)
	}
}

func TestRenderIntentStoresStructuredPayload(t *testing.T) {
	intent := RenderIntent{
		Kind:    RenderIntentTable,
		Title:   "Models",
		Columns: []string{"Provider", "Model"},
		Rows:    []RenderTableRow{{Cells: []string{"openai", "gpt-4o"}}},
	}
	if len(intent.Rows) != 1 || len(intent.Rows[0].Cells) != 2 {
		t.Fatalf("unexpected table payload: %#v", intent)
	}
}

func TestRenderIntentStoresRichListPayloads(t *testing.T) {
	intent := RenderIntent{
		Kind: RenderIntentActionList,
		Actions: []RenderAction{{
			Label:   "Inspect diagnostics",
			Command: "/status diagnostics",
			Detail:  "Show corrective loop output",
			Status:  "recommended",
		}},
		Options: []RenderOption{{
			Label:    "anthropic/claude-sonnet-4",
			Detail:   "ctx=200000 caps=text,tools",
			Status:   "ready",
			Hint:     "/model anthropic/claude-sonnet-4",
			Selected: true,
		}},
		DetailRows: []RenderDetailRow{{
			Label:  "Provider",
			Value:  "anthropic",
			Status: "ok",
			Detail: "Authenticated and ready",
		}},
	}
	if len(intent.Actions) != 1 || intent.Actions[0].Command != "/status diagnostics" {
		t.Fatalf("unexpected action payload: %#v", intent)
	}
	if len(intent.Options) != 1 || !intent.Options[0].Selected {
		t.Fatalf("unexpected option payload: %#v", intent)
	}
	if len(intent.DetailRows) != 1 || intent.DetailRows[0].Label != "Provider" {
		t.Fatalf("unexpected detail rows payload: %#v", intent)
	}
}

func TestRenderIntentHasContent(t *testing.T) {
	if (RenderIntent{Kind: RenderIntentSummaryCard}).HasContent() {
		t.Fatalf("expected empty intent to report no content")
	}
	if !(RenderIntent{Kind: RenderIntentSummaryCard, Title: "Status"}).HasContent() {
		t.Fatalf("expected title intent to report content")
	}
	if !(RenderIntent{Kind: RenderIntentDetailRows, DetailRows: []RenderDetailRow{{Label: "Mode", Value: "auto"}}}).HasContent() {
		t.Fatalf("expected structured rows intent to report content")
	}
}

func TestRenderIntentCloneDeepCopiesSlices(t *testing.T) {
	orig := RenderIntent{
		Kind:    RenderIntentTable,
		Title:   "Models",
		Columns: []string{"Provider", "Model"},
		Rows:    []RenderTableRow{{Cells: []string{"openai", "gpt-4o"}}},
	}
	clone := orig.Clone()
	clone.Columns[0] = "X"
	clone.Rows[0].Cells[0] = "Y"
	if orig.Columns[0] != "Provider" {
		t.Fatalf("expected original columns unchanged, got %#v", orig.Columns)
	}
	if orig.Rows[0].Cells[0] != "openai" {
		t.Fatalf("expected original row cells unchanged, got %#v", orig.Rows)
	}
}
