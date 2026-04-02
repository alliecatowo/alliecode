package tui

import "testing"

func TestWave5QuickOpenRankingUsesTokenCoverage(t *testing.T) {
	state := newQuickOpenState([]quickOpenItem{
		{label: "Permission queue", detail: "Inspect live approvals", value: "workflow.permission", keywords: "permission queue approvals"},
		{label: "Reference picker", detail: "Use path autocomplete", value: "workflow.references", keywords: "files context"},
	})
	state.setQuery("permission approvals")
	item, ok := state.selectedItem()
	if !ok || item.value != "workflow.permission" {
		t.Fatalf("expected token-rich permission workflow first, got ok=%t item=%+v", ok, item)
	}
}
