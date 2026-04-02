package state

import (
	"path/filepath"
	"testing"
)

func TestSessionMetadataRepositoryQuery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "session-metadata.json")
	repo := NewSessionMetadataRepository(path)

	_, _ = repo.Upsert(SessionMetadataUpdate{SessionID: "a1", ProjectPath: "/repo/a", Title: "alpha", UpdatedAt: 10, MessageCount: 1, RuntimeSurface: "cli", CommandSurface: "repl"})
	_, _ = repo.Upsert(SessionMetadataUpdate{SessionID: "b2", ProjectPath: "/repo/b", Title: "beta", UpdatedAt: 20, MessageCount: 3, RuntimeSurface: "cli", CommandSurface: "batch"})

	items, err := repo.Query(SessionMetadataQuery{ProjectContains: "/repo/b", MinMessages: 2, SortByMessages: true, Limit: 5})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(items) != 1 || items[0].SessionID != "b2" {
		t.Fatalf("unexpected query result: %+v", items)
	}
}
