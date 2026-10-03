package tui

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRenderIntentPlainDelegatesToNodePipeline(t *testing.T) {
	intent := types.RenderIntent{Kind: types.RenderIntentDetailRows, Title: "Token", DetailRows: []types.RenderDetailRow{{Label: "Set", Value: "yes", Status: "ok"}}}
	got := renderIntentPlain(intent)
	want := "Token\nSet: yes [ok]"
	if got != want {
		t.Fatalf("unexpected render from node pipeline\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
