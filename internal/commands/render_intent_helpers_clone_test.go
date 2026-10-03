package commands

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestCloneRenderIntentsDropsEmptyAndClonesNested(t *testing.T) {
	src := []types.RenderIntent{
		{Kind: types.RenderIntentSummaryCard},
		{Kind: types.RenderIntentTable, Title: "Models", Columns: []string{"Provider"}, Rows: []types.RenderTableRow{{Cells: []string{"openai"}}}},
	}
	got := cloneRenderIntents(src)
	if len(got) != 1 {
		t.Fatalf("expected one non-empty intent, got %d", len(got))
	}
	got[0].Columns[0] = "X"
	got[0].Rows[0].Cells[0] = "Y"
	if src[1].Columns[0] != "Provider" {
		t.Fatalf("expected source columns unchanged, got %#v", src[1].Columns)
	}
	if src[1].Rows[0].Cells[0] != "openai" {
		t.Fatalf("expected source rows unchanged, got %#v", src[1].Rows)
	}
}
