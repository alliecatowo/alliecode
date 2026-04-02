package history

import (
	"testing"
	"time"
)

func TestQueryEventsPageRequireModelAndDisplayContains(t *testing.T) {
	store, err := NewJSONLStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}
	_ = store.Append(Event{Type: EventCustom, SessionID: "s1", Project: "/repo/a", Display: "deploy alpha", Model: "gpt", Timestamp: time.Unix(1, 0).UTC()})
	_ = store.Append(Event{Type: EventCustom, SessionID: "s2", Project: "/repo/a", Display: "deploy beta", Timestamp: time.Unix(2, 0).UTC()})
	repo := NewRepository(store)
	page, err := repo.QueryEventsPage(EventPageQuery{HistoryQuery: HistoryQuery{Project: "/repo/a"}, RequireModel: true, DisplayContains: "alpha"})
	if err != nil {
		t.Fatalf("QueryEventsPage() error = %v", err)
	}
	if page.Total != 1 || page.Items[0].SessionID != "s1" {
		t.Fatalf("unexpected page: %+v", page)
	}
}
