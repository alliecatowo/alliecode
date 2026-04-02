package migrations

import "testing"

func TestBuildRunReport(t *testing.T) {
	report := BuildRunReport(RunDiagnostics{AppliedVersions: []string{"1", "2"}, SkippedVersions: []string{"3"}, FailedVersion: "", DurationMs: 42, AppliedCount: 2, SkippedCount: 1})
	if report.LastApplied != "2" || report.HasFailure {
		t.Fatalf("unexpected report: %+v", report)
	}
}
