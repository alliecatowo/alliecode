package session

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRepositoryQueryPageAndSummary(t *testing.T) {
	root := t.TempDir()
	repo := NewRepository(root)
	store, err := repo.Create("s1", SessionStartOptions{ProjectPath: "/repo/a", RuntimeSurface: "cli", Provider: "openai"})
	if err != nil {
		t.Fatalf("Create(s1) error = %v", err)
	}
	if err := store.AppendMessage(types.NewTextMessage(types.RoleAssistant, "hi")); err != nil {
		t.Fatalf("AppendMessage() error = %v", err)
	}
	store2, err := repo.Create("s2", SessionStartOptions{ProjectPath: "/repo/a", RuntimeSurface: "ide", Provider: "anthropic"})
	if err != nil {
		t.Fatalf("Create(s2) error = %v", err)
	}
	if err := store2.AppendMessage(types.NewTextMessage(types.RoleUser, "yo")); err != nil {
		t.Fatalf("AppendMessage() error = %v", err)
	}

	page, err := repo.QueryPage(PageQuery{Query: Query{ProjectContains: "/repo", Limit: 0}, LastRoleContains: "user", Offset: 0, Limit: 1})
	if err != nil {
		t.Fatalf("QueryPage() error = %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}

	summary, err := repo.QuerySummary(PageQuery{Query: Query{ProjectContains: "/repo", Limit: 0}, RequireRuntime: true, RequireProvider: true})
	if err != nil {
		t.Fatalf("QuerySummary() error = %v", err)
	}
	if summary.Total != 2 || summary.ProviderCount["openai"] != 1 || summary.DistinctSessions != 2 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}
