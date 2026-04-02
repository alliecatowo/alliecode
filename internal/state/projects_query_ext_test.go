package state

import (
	"path/filepath"
	"testing"
)

func TestProjectsRepositoryQueryPageAndSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects.json")
	repo := NewProjectsRepository(path)
	registry := ProjectsRegistry{Projects: map[string]ProjectMetadata{
		"/repo/a": {LastSessionID: "s1", OpenCount: 5, DistinctSessionCount: 2, LastOpenedAt: 30, LastProvider: "openai", LastRuntimeSurface: "cli", LastCommandSurface: "repl"},
		"/repo/b": {LastSessionID: "s2", OpenCount: 1, DistinctSessionCount: 1, LastOpenedAt: 20, LastProvider: "anthropic", LastRuntimeSurface: "ide", LastCommandSurface: "panel"},
		"/repo/c": {OpenCount: 2, DistinctSessionCount: 1, LastOpenedAt: 10},
	}}
	if err := repo.Save(registry); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	page, err := repo.QueryPage(ProjectPageQuery{ProjectQuery: ProjectQuery{Limit: 0}, PathContainsAny: []string{"repo"}, MinOpenCount: 1, Offset: 1, Limit: 1, RequireSession: true})
	if err != nil {
		t.Fatalf("QueryPage() error = %v", err)
	}
	if page.Total != 2 || len(page.Items) != 1 || page.HasMore {
		t.Fatalf("unexpected page: %+v", page)
	}

	summary, err := repo.QuerySummary(ProjectPageQuery{ProjectQuery: ProjectQuery{Limit: 0}, RequireSession: true})
	if err != nil {
		t.Fatalf("QuerySummary() error = %v", err)
	}
	if summary.Total != 2 || summary.ProviderCounts["openai"] != 1 || summary.ProjectsWithSession != 2 || summary.MostRecentOpenedAt != 30 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestProjectsRepositoryQueryIndexPageOffset(t *testing.T) {
	path := filepath.Join(t.TempDir(), "projects-index.json")
	repo := NewProjectsRepository(path)
	if err := repo.Save(ProjectsRegistry{Projects: map[string]ProjectMetadata{
		"/repo/a": {OpenCount: 3, LastOpenedAt: 3},
		"/repo/b": {OpenCount: 2, LastOpenedAt: 2},
		"/repo/c": {OpenCount: 1, LastOpenedAt: 1},
	}}); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	page, err := repo.QueryIndexPage(ProjectIndexQuery{Offset: 1, Limit: 1})
	if err != nil {
		t.Fatalf("QueryIndexPage() error = %v", err)
	}
	if page.Total != 3 || len(page.Items) != 1 || page.Items[0].Path != "/repo/b" {
		t.Fatalf("unexpected index page: %+v", page)
	}
}
