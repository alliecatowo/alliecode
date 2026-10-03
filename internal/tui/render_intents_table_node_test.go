package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRenderIntentsPlainTableViaNodes(t *testing.T) {
	got := renderIntentsPlain([]types.RenderIntent{{Kind: types.RenderIntentTable, Title: "Models", Columns: []string{"Provider", "Model"}, Rows: []types.RenderTableRow{{Cells: []string{"openai", "gpt-4o"}}}}})
	for _, token := range []string{"Models", "Provider | Model", "openai   | gpt-4o"} {
		if !strings.Contains(got, token) {
			t.Fatalf("expected %q in %q", token, got)
		}
	}
}
