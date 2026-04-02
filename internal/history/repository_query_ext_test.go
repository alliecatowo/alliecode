package history

import (
	"testing"
	"time"
)

func TestRepositoryQueryEventsPageAndSummary(t *testing.T) {
	root := t.TempDir()
	store, err := NewJSONLStore(root)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}
	if err := store.Append(Event{Type: EventCustom, SessionID: "s1", Project: "/repo/a", Display: "ship build", Provider: "openai", Model: "gpt-4o-mini", RuntimeSurface: "cli", CommandSurface: "repl", Tags: []string{"prod"}, Timestamp: time.Unix(10, 0).UTC()}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	if err := store.Append(Event{Type: EventCustom, SessionID: "s2", Project: "/repo/a", Display: "test run", Provider: "anthropic", Model: "claude", RuntimeSurface: "ide", CommandSurface: "panel", Tags: []string{"ci"}, Timestamp: time.Unix(20, 0).UTC()}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	repo := NewRepository(store)

	withTags := true
	page, err := repo.QueryEventsPage(EventPageQuery{HistoryQuery: HistoryQuery{Project: "/repo/a"}, Provider: "anthropic", ProjectContains: "/repo", DisplayAny: []string{"run"}, HasTags: &withTags, Offset: 0, Limit: 1})
	if err != nil {
		t.Fatalf("QueryEventsPage() error = %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].SessionID != "s2" {
		t.Fatalf("unexpected page: %+v", page)
	}

	summary, err := repo.QueryEventsSummary(EventPageQuery{HistoryQuery: HistoryQuery{Project: "/repo/a"}})
	if err != nil {
		t.Fatalf("QueryEventsSummary() error = %v", err)
	}
	if summary.Total != 2 || summary.ByProvider["openai"] != 1 || summary.DistinctUsers != 2 || summary.ByCommand["repl"] != 1 || summary.DistinctProjects != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}
