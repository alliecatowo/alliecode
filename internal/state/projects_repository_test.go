package state

import (
	"path/filepath"
	"testing"
	"time"
)

func TestProjectsRepositoryRecordAndQuery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects-state.json")
	repo := NewProjectsRepository(path)
	now := time.Now().UTC()

	_, err := repo.RecordOpen(ProjectOpenRecord{ProjectPath: "/r/a", SessionID: "s1", SessionTitle: "build", OpenedAt: now.Add(-time.Minute)})
	if err != nil {
		t.Fatalf("RecordOpen(a) error = %v", err)
	}
	_, err = repo.RecordOpen(ProjectOpenRecord{ProjectPath: "/r/b", SessionID: "s2", SessionTitle: "deploy", OpenedAt: now})
	if err != nil {
		t.Fatalf("RecordOpen(b) error = %v", err)
	}

	items, err := repo.Query(ProjectQuery{ContainsTitle: "dep", Limit: 10})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(items) != 1 || items[0].Path != "/r/b" {
		t.Fatalf("unexpected query result: %+v", items)
	}
}
