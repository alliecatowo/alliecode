package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRenderIntentsPlainSummaryCard(t *testing.T) {
	got := renderIntentsPlain([]types.RenderIntent{{
		Kind:    types.RenderIntentSummaryCard,
		Title:   "Status",
		Summary: "Runtime summary.",
		Fields:  []types.RenderField{{Label: "Model", Value: "gpt-4o"}},
	}})
	for _, want := range []string{"Status", "Runtime summary.", "Model: gpt-4o"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in %q", want, got)
		}
	}
}

func TestRenderIntentsPlainTable(t *testing.T) {
	got := renderIntentsPlain([]types.RenderIntent{{
		Kind:    types.RenderIntentTable,
		Title:   "Models",
		Columns: []string{"Provider", "Model"},
		Rows:    []types.RenderTableRow{{Cells: []string{"openai", "gpt-4o"}}},
	}})
	for _, want := range []string{"Models", "Provider | Model", "openai   | gpt-4o"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in %q", want, got)
		}
	}
}

func TestRenderIntentsPlainDetailRowsAndActions(t *testing.T) {
	got := renderIntentsPlain([]types.RenderIntent{
		{Kind: types.RenderIntentDetailRows, Title: "Session token", DetailRows: []types.RenderDetailRow{{Label: "Set", Value: "yes", Status: "set", Detail: "Token exists"}}},
		{Kind: types.RenderIntentActionList, Title: "Actions", Actions: []types.RenderAction{{Label: "Show session status", Command: "/session status", Status: "recommended", Detail: "Inspect transport health"}}},
	})
	for _, want := range []string{"Session token", "Set: yes [set] - Token exists", "Show session status: /session status [recommended] - Inspect transport health"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in %q", want, got)
		}
	}
}

func TestRenderIntentsPlainOptionList(t *testing.T) {
	got := renderIntentsPlain([]types.RenderIntent{{
		Kind:    types.RenderIntentOptionList,
		Title:   "Matching commands",
		Summary: "Structured command list.",
		Options: []types.RenderOption{{Label: "/status", Detail: "Show runtime status", Status: "general", Hint: "/status", Selected: true}},
	}})
	for _, want := range []string{"Matching commands", "Structured command list.", "* /status [general]: Show runtime status -> /status"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected %q in %q", want, got)
		}
	}
}
