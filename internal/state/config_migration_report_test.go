package state

import "testing"

func TestBuildConfigMigrationReports(t *testing.T) {
	states := []ConfigDiagnosticsState{
		{LastConfigPath: "/c/a", LastProjectPath: "/repo/a", LayeringMode: "layered", LastMigrationVersion: "3", MigrationAppliedCount: 2, MigrationSkippedCount: 1, LastValidationErrorCount: 1, LastValidationPassed: true, LastValidatedAtUnix: 30},
		{LastConfigPath: "/c/b", LastProjectPath: "/repo/b", LastMigrationVersion: "2", MigrationAppliedCount: 0, MigrationSkippedCount: 3, LastValidationPassed: false, LastValidatedAtUnix: 20},
	}
	out := BuildConfigMigrationReports(states, ConfigMigrationQuery{LayeringMode: "layered", MigrationVersion: "3", MinApplied: 1, MinErrors: 1, Limit: 5})
	if len(out) != 1 || out[0].ConfigPath != "/c/a" || out[0].LayeringMode != "layered" {
		t.Fatalf("unexpected migration reports: %+v", out)
	}
}
