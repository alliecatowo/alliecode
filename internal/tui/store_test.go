package tui

import "testing"

func TestNewTUIStoreDefaults(t *testing.T) {
	store := newTUIStore()
	if store.runtime.state != stateIdle {
		t.Fatalf("expected idle runtime state, got %d", store.runtime.state)
	}
	if store.runtime.inputMode != inputModeChat {
		t.Fatalf("expected chat input mode, got %s", store.runtime.inputMode)
	}
	if store.search.mode != searchModeTimeline {
		t.Fatalf("expected timeline search mode, got %d", store.search.mode)
	}
	if store.search.query != "" {
		t.Fatalf("expected empty search query, got %q", store.search.query)
	}
}

func TestSyncStoreFromLegacyCopiesSearchAndRuntimeSlices(t *testing.T) {
	app := New(Config{})
	app.state = stateSearch
	app.inputMode = inputModeQuickOpen
	app.searchMode = searchModeQuickOpen
	app.searchQuery = "model"
	app.searchTimelineQuery = "timeline"
	app.searchQuickOpenQuery = "quick"
	app.searchHistoryQuery = "history"
	app.matchPos = 3

	app.syncStoreFromLegacy()

	if app.store.runtime.state != stateSearch {
		t.Fatalf("expected runtime state search, got %d", app.store.runtime.state)
	}
	if app.store.runtime.inputMode != inputModeQuickOpen {
		t.Fatalf("expected runtime input mode quick-open, got %s", app.store.runtime.inputMode)
	}
	if app.store.search.mode != searchModeQuickOpen {
		t.Fatalf("expected search mode quick-open, got %d", app.store.search.mode)
	}
	if app.store.search.query != "model" {
		t.Fatalf("expected query copied into store, got %q", app.store.search.query)
	}
	if app.store.search.timelineQuery != "timeline" || app.store.search.quickOpenQuery != "quick" || app.store.search.historyQuery != "history" {
		t.Fatalf("expected all search query slices copied, got timeline=%q quick=%q history=%q", app.store.search.timelineQuery, app.store.search.quickOpenQuery, app.store.search.historyQuery)
	}
	if app.store.search.matchPos != 3 {
		t.Fatalf("expected match position copied, got %d", app.store.search.matchPos)
	}
}
