package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLayerRepositoryQueryDiagnosticsPageAndSummary(t *testing.T) {
	project := t.TempDir()
	path := filepath.Join(project, ".alliecode", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte("default_provider: openai\ndefault_model: gpt-4o-mini\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	repo := NewLayerRepository(project)
	page, err := repo.QueryDiagnosticsPage(LayerPageQuery{Limit: 5})
	if err != nil {
		t.Fatalf("QueryDiagnosticsPage() error = %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}
	summary, err := repo.QueryDiagnosticsSummary(LayerPageQuery{})
	if err != nil {
		t.Fatalf("QueryDiagnosticsSummary() error = %v", err)
	}
	if summary.Total != 1 || summary.StrictHydration < 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}
