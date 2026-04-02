package session

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRepositoryCreateLoadAndQuery(t *testing.T) {
	root := t.TempDir()
	repo := NewRepository(root)
	store, err := repo.Create("s1", SessionStartOptions{ProjectPath: "/repo/a", RuntimeSurface: "cli", Provider: "openai", Model: "gpt-4o-mini"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := store.AppendMessage(types.NewTextMessage(types.RoleUser, "hi")); err != nil {
		t.Fatalf("AppendMessage() error = %v", err)
	}

	tr, err := repo.Load("s1")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if tr.SessionID != "s1" {
		t.Fatalf("unexpected transcript: %+v", tr)
	}

	items, err := repo.Query(Query{ProjectContains: "/repo/a", Provider: "openai"})
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(items) != 1 || items[0].SessionID != "s1" {
		t.Fatalf("unexpected query result: %+v", items)
	}
}
