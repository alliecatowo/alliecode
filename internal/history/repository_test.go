package history

import (
	"testing"
	"time"
)

func TestRepositoryBuildIndex(t *testing.T) {
	root := t.TempDir()
	store, err := NewJSONLStore(root)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}
	if err := store.Append(Event{Type: EventCustom, SessionID: "s1", Project: "/repo/a", Display: "ship", Timestamp: time.Unix(1, 0).UTC()}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	repo := NewRepository(store)
	items, err := repo.BuildIndex(HistoryQuery{Project: "/repo/a", MaxItems: 10}, IndexQuery{DisplayContains: "shi"})
	if err != nil {
		t.Fatalf("BuildIndex() error = %v", err)
	}
	if len(items) != 1 || items[0].Display != "ship" {
		t.Fatalf("unexpected index result: %+v", items)
	}
}
