package tui

import "testing"

func TestWave5QuickOpenRankingPrefersExactValue(t *testing.T) {
	state := newQuickOpenState([]quickOpenItem{
		{label: "History search", value: "search.history", keywords: "search prompt"},
		{label: "Timeline", value: "search.timeline", keywords: "search rows"},
	})
	state.setQuery("search.history")
	item, ok := state.selectedItem()
	if !ok || item.value != "search.history" {
		t.Fatalf("expected exact value match selected first, got ok=%t item=%+v", ok, item)
	}
}
