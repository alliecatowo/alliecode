package migrations

import (
	"testing"
	"time"
)

func TestBuildReportAppliedVersionsOnlyAppliedStatus(t *testing.T) {
	report := BuildReport([]LedgerEntry{
		{Version: "v1", Status: "applied", AppliedAt: time.Unix(1, 0).UTC()},
		{Version: "v2", Status: "failed", AppliedAt: time.Unix(2, 0).UTC()},
	})
	if len(report.AppliedVersions) != 1 || report.AppliedVersions[0] != "v1" {
		t.Fatalf("unexpected applied versions: %+v", report.AppliedVersions)
	}
}
