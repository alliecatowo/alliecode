package state

import (
	"path/filepath"
	"testing"
)

func TestSessionMetadataRepositoryQueryPageAndSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sessions-meta.json")
	repo := NewSessionMetadataRepository(path)
	registry := SessionMetadataRegistry{Sessions: map[string]SessionMetadata{
		"s1": {ProjectPath: "/repo/a", MessageCount: 4, EventCount: 8, LastRole: "assistant", RuntimeSurface: "cli", Summary: "a"},
		"s2": {ProjectPath: "/repo/a", MessageCount: 1, EventCount: 2, LastRole: "user", RuntimeSurface: "ide", Summary: ""},
	}}
	if err := repo.Save(registry); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	hasRuntime := true
	page, err := repo.QueryPage(SessionMetadataPageQuery{SessionMetadataQuery: SessionMetadataQuery{Limit: 0}, ProjectAny: []string{"/repo"}, HasRuntime: &hasRuntime, LastRole: "assistant", Offset: 0, Limit: 1})
	if err != nil {
		t.Fatalf("QueryPage() error = %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].SessionID != "s1" {
		t.Fatalf("unexpected page: %+v", page)
	}

	summary, err := repo.QuerySummary(SessionMetadataPageQuery{SessionMetadataQuery: SessionMetadataQuery{Limit: 0}, MinEvents: 2})
	if err != nil {
		t.Fatalf("QuerySummary() error = %v", err)
	}
	if summary.Total != 2 || summary.TotalEvents != 10 || summary.RoleCounts["assistant"] != 1 || summary.DistinctProjects != 1 || summary.SessionsWithSummary != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}
