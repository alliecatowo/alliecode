package tui

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestBuildIntentRenderTreeSkipsEmptyIntent(t *testing.T) {
	tree := buildIntentRenderTree([]types.RenderIntent{{Kind: types.RenderIntentSummaryCard}, {
		Kind:   types.RenderIntentSummaryCard,
		Title:  "Status",
		Fields: []types.RenderField{{Label: "Model", Value: "gpt-4o"}},
	}})
	if len(tree.Children) != 1 {
		t.Fatalf("expected one non-empty section, got %d", len(tree.Children))
	}
}

func TestBuildIntentNodeOptionList(t *testing.T) {
	node := buildIntentNode(types.RenderIntent{
		Kind:    types.RenderIntentOptionList,
		Title:   "Commands",
		Summary: "Structured",
		Options: []types.RenderOption{{Label: "/status", Selected: true, Status: "general", Detail: "Show state", Hint: "/status"}},
	})
	if len(node.Children) < 3 {
		t.Fatalf("expected header and option children, got %d", len(node.Children))
	}
}
