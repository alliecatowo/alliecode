package session

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRepositoryQueryExtendedFiltersAndSnapshots(t *testing.T) {
	root := t.TempDir()
	repo := NewRepository(root)
	store, err := repo.Create("s1", SessionStartOptions{ProjectPath: "/repo/a", RuntimeSurface: "cli", CommandSurface: "repl", Provider: "openai", Model: "gpt"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := store.AppendMessage(types.NewTextMessage(types.RoleAssistant, "ok")); err != nil {
		t.Fatalf("AppendMessage() error = %v", err)
	}

	out, err := repo.Query(Query{CommandSurface: "repl", LastRole: "assistant", MinMessages: 1, Limit: 2})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(out) != 1 || out[0].SessionID != "s1" {
		t.Fatalf("unexpected query output: %+v", out)
	}

	snaps, err := repo.QuerySnapshots(Query{SessionContains: "s1"}, SnapshotQuery{Provider: "openai", Limit: 1})
	if err != nil {
		t.Fatalf("QuerySnapshots() error = %v", err)
	}
	if len(snaps) != 1 || snaps[0].SessionID != "s1" {
		t.Fatalf("unexpected snapshots output: %+v", snaps)
	}
}

func TestRepositoryQueryPageOffset(t *testing.T) {
	root := t.TempDir()
	repo := NewRepository(root)
	if _, err := repo.Create("s1", SessionStartOptions{ProjectPath: "/repo/a", RuntimeSurface: "cli"}); err != nil {
		t.Fatalf("Create(s1) error = %v", err)
	}
	if _, err := repo.Create("s2", SessionStartOptions{ProjectPath: "/repo/a", RuntimeSurface: "cli"}); err != nil {
		t.Fatalf("Create(s2) error = %v", err)
	}
	page, err := repo.QueryPage(PageQuery{Query: Query{ProjectContains: "/repo", Limit: 0}, RequireModel: false, Offset: 1, Limit: 1})
	if err != nil {
		t.Fatalf("QueryPage() error = %v", err)
	}
	if page.Total != 2 || len(page.Items) != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}
}
