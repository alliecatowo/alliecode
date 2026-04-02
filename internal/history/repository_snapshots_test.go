package history

import (
	"testing"
	"time"
)

func TestRepositoryQuerySnapshots(t *testing.T) {
	root := t.TempDir()
	store, err := NewJSONLStore(root)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}
	if err := store.Append(Event{Type: EventCustom, SessionID: "s1", Project: "/repo/a", Display: "deploy", Timestamp: time.Unix(10, 0).UTC()}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	repo := NewRepository(store)
	out, err := repo.QuerySnapshots(HistoryQuery{Project: "/repo/a", MaxItems: 10}, SnapshotQuery{DisplayContains: "dep", Limit: 5})
	if err != nil {
		t.Fatalf("QuerySnapshots() error = %v", err)
	}
	if len(out) != 1 || out[0].Display != "deploy" {
		t.Fatalf("unexpected snapshots: %+v", out)
	}
}

func TestRepositoryQueryEventsSummary(t *testing.T) {
	root := t.TempDir()
	store, err := NewJSONLStore(root)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}
	if err := store.Append(Event{Type: EventCustom, SessionID: "s1", Project: "/repo/a", Display: "deploy", Provider: "openai", Timestamp: time.Unix(10, 0).UTC()}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	repo := NewRepository(store)
	summary, err := repo.QueryEventsSummary(EventPageQuery{HistoryQuery: HistoryQuery{Project: "/repo/a"}})
	if err != nil {
		t.Fatalf("QueryEventsSummary() error = %v", err)
	}
	if summary.Total != 1 || summary.ByProvider["openai"] != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}
