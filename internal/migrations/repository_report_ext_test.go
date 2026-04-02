package migrations

import (
	"path/filepath"
	"testing"
	"time"
)

func TestBuildReportAndQueryPage(t *testing.T) {
	entries := []LedgerEntry{
		{Version: "v1", Status: "applied", DurationMs: 10, AppliedAt: time.Unix(1, 0).UTC()},
		{Version: "v2", Status: "failed", DurationMs: 30, AppliedAt: time.Unix(2, 0).UTC()},
		{Version: "v3", Status: "applied", DurationMs: 20, AppliedAt: time.Unix(3, 0).UTC()},
	}
	report := BuildReport(entries)
	if report.Total != 3 || report.LastVersion != "v3" || report.Durations["avg_ms"] != 20 || len(report.AppliedVersions) != 2 {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestRepositoryQueryReportAndPage(t *testing.T) {
	ledger := NewLedger(filepath.Join(t.TempDir(), "ledger.jsonl"))
	if err := ledger.RecordDetailed(LedgerEntry{Version: "v1", Status: "applied", DurationMs: 10, AppliedAt: time.Unix(1, 0).UTC()}); err != nil {
		t.Fatalf("RecordDetailed(v1) error = %v", err)
	}
	if err := ledger.RecordDetailed(LedgerEntry{Version: "v2", Status: "failed", DurationMs: 30, AppliedAt: time.Unix(2, 0).UTC()}); err != nil {
		t.Fatalf("RecordDetailed(v2) error = %v", err)
	}
	repo := NewRepository(ledger)
	report, err := repo.QueryReport(Query{})
	if err != nil {
		t.Fatalf("QueryReport() error = %v", err)
	}
	if report.Total != 2 || report.Statuses["failed"] != 1 {
		t.Fatalf("unexpected report: %+v", report)
	}
	page, err := repo.QueryPage(Query{}, 1)
	if err != nil {
		t.Fatalf("QueryPage() error = %v", err)
	}
	if page.Total != 2 || len(page.Items) != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}
}
