package migrations

import (
	"testing"
	"time"
)

func TestFilterEntriesExtendedAndSummary(t *testing.T) {
	entries := []LedgerEntry{
		{Version: "20260401_001", Description: "seed defaults", Status: "applied", DurationMs: 10, AppliedAt: time.Unix(1, 0).UTC()},
		{Version: "20260401_002", Description: "skip patch", Status: "skipped", DurationMs: 5, AppliedAt: time.Unix(2, 0).UTC()},
	}
	out := FilterEntries(entries, Query{VersionSuffix: "002", DescriptionContains: "skip", StatusIn: []string{"skipped"}, MinDurationMs: 1, Limit: 5})
	if len(out) != 1 || out[0].Version != "20260401_002" {
		t.Fatalf("unexpected filter output: %+v", out)
	}
	summary := SummarizeEntries(entries)
	if summary.Total != 2 || summary.Applied != 1 || summary.Skipped != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestBuildReportAggregatesStatusAndDurations(t *testing.T) {
	entries := []LedgerEntry{{Version: "1", Status: "applied", DurationMs: 10, AppliedAt: time.Unix(1, 0).UTC()}, {Version: "2", Status: "failed", DurationMs: 20, AppliedAt: time.Unix(2, 0).UTC()}}
	report := BuildReport(entries)
	if report.Total != 2 || report.Statuses["failed"] != 1 || report.Durations["max_ms"] != 20 || len(report.AppliedVersions) != 1 {
		t.Fatalf("unexpected report: %+v", report)
	}
}
