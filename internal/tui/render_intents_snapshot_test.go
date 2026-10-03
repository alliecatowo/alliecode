package tui

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRenderIntentsSnapshotComposite(t *testing.T) {
	got := renderIntentsPlain([]types.RenderIntent{
		{Kind: types.RenderIntentSummaryCard, Title: "Session info", Fields: []types.RenderField{{Label: "Hosted", Value: "yes"}, {Label: "Connected", Value: "no"}}},
		{Kind: types.RenderIntentDetailRows, Title: "Token", DetailRows: []types.RenderDetailRow{{Label: "Set", Value: "yes", Status: "set", Detail: "Prefix available"}}},
		{Kind: types.RenderIntentActionList, Title: "Actions", Actions: []types.RenderAction{{Label: "Diagnostics", Command: "/session diagnostics", Status: "recommended", Detail: "Inspect reconnect health"}}},
	})
	const want = "Session info\nHosted: yes\nConnected: no\n\nToken\nSet: yes [set] - Prefix available\n\nActions\n- Diagnostics: /session diagnostics [recommended] - Inspect reconnect health"
	if got != want {
		t.Fatalf("snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
