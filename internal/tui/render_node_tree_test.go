package tui

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRenderNodeTreePlainSectionAndLines(t *testing.T) {
	root := types.RenderNode{Kind: types.RenderNodeRoot, Children: []types.RenderNode{{
		Kind: types.RenderNodeSection,
		Children: []types.RenderNode{
			{Kind: types.RenderNodeLine, Text: "Status"},
			{Kind: types.RenderNodeLine, Text: "Model: gpt-4o"},
		},
	}}}
	got := renderNodeTreePlain(root)
	want := "Status\nModel: gpt-4o"
	if got != want {
		t.Fatalf("unexpected node tree render\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestRenderNodeTreePlainTable(t *testing.T) {
	root := types.RenderNode{Kind: types.RenderNodeRoot, Children: []types.RenderNode{{
		Kind: types.RenderNodeSection,
		Children: []types.RenderNode{{
			Kind: types.RenderNodeTable,
			Children: []types.RenderNode{
				{Kind: types.RenderNodeRow, Cells: []string{"Provider", "Model"}},
				{Kind: types.RenderNodeRow, Cells: []string{"openai", "gpt-4o"}},
			},
		}},
	}}}
	got := renderNodeTreePlain(root)
	if got == "" || got[:8] != "Provider" {
		t.Fatalf("expected table output, got %q", got)
	}
}
