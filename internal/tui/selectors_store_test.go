package tui

import "testing"

func TestSearchSelectorsReflectMutations(t *testing.T) {
	app := New(Config{})

	app.setSearchModeValue(searchModeHistory)
	app.setSearchQueryValue("hello")
	app.setSearchMatchPosValue(2)
	app.setSearchTimelineQueryValue("t")
	app.setSearchQuickOpenQueryValue("q")
	app.setSearchHistoryQueryValue("h")

	if app.searchModeValue() != searchModeHistory {
		t.Fatalf("expected history mode, got %d", app.searchModeValue())
	}
	if app.searchQueryValue() != "hello" {
		t.Fatalf("expected query hello, got %q", app.searchQueryValue())
	}
	if app.searchMatchPosValue() != 2 {
		t.Fatalf("expected match index 2, got %d", app.searchMatchPosValue())
	}
	if app.searchTimelineQuery != "t" || app.searchQuickOpenQuery != "q" || app.searchHistoryQuery != "h" {
		t.Fatalf("expected mirrored query caches set")
	}
}

func TestStateSelectorsDriveSearchActivity(t *testing.T) {
	app := New(Config{})
	app.setStateValue(stateSearch)
	app.setSearchModeValue(searchModeTimeline)
	if !app.timelineSearchActive() {
		t.Fatalf("expected timeline search active")
	}
	app.setSearchModeValue(searchModeQuickOpen)
	if !app.quickOpenSearchActive() {
		t.Fatalf("expected quick-open search active")
	}
	app.setSearchModeValue(searchModeHistory)
	if !app.historySearchActive() {
		t.Fatalf("expected history search active")
	}
	app.setStateValue(stateIdle)
	if app.timelineSearchActive() || app.quickOpenSearchActive() || app.historySearchActive() {
		t.Fatalf("expected no search selector active outside search state")
	}
}
