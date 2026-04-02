package history

import (
	"testing"
	"time"
)

func TestRepositoryBuildIndexPageAndSummary(t *testing.T) {
	root := t.TempDir()
	store, err := NewJSONLStore(root)
	if err != nil {
		t.Fatalf("NewJSONLStore() error = %v", err)
	}
	if err := store.Append(Event{Type: EventCustom, SessionID: "s1", Project: "/repo/a", Display: "a", RuntimeSurface: "cli", CommandSurface: "repl", Timestamp: time.Unix(1, 0).UTC()}); err != nil {
		t.Fatalf("Append(custom) error = %v", err)
	}
	if err := store.Append(Event{Type: EventMessage, SessionID: "s2", Project: "/repo/a", Display: "b", RuntimeSurface: "ide", CommandSurface: "panel", Timestamp: time.Unix(2, 0).UTC()}); err != nil {
		t.Fatalf("Append(message) error = %v", err)
	}

	repo := NewRepository(store)
	page, err := repo.BuildIndexPage(HistoryQuery{Project: "/repo/a"}, IndexQuery{}, 0, 1)
	if err != nil {
		t.Fatalf("BuildIndexPage() error = %v", err)
	}
	if page.Total != 2 || len(page.Items) != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}

	sum, err := repo.BuildIndexSummary(HistoryQuery{Project: "/repo/a"}, IndexQuery{})
	if err != nil {
		t.Fatalf("BuildIndexSummary() error = %v", err)
	}
	if sum.Total != 2 || sum.TypeCounts[string(EventCustom)] != 1 || sum.ByRuntime["cli"] != 1 || sum.DistinctSessions != 2 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
}
