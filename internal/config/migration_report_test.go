package config

import "testing"

func TestBuildAndQueryMigrationReports(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Startup.HydrationMode = "strict"
	cfg.Startup.StrictHydration = true
	cfg.Remote.Mode = "local"
	cfg.MigrationVersion = 3
	report := BuildMigrationReport(cfg, t.TempDir(), []string{"a", "b"})
	strict := true
	out := QueryMigrationReports([]MigrationReport{report}, MigrationReportQuery{HydrationMode: "strict", RemoteMode: "local", StrictHydration: &strict, Provider: "openai", MinVersion: 2, MinApplied: 1, Limit: 1})
	if len(out) != 1 || out[0].MigrationVersion != 3 || len(out[0].AppliedPatches) != 2 {
		t.Fatalf("unexpected migration report output: %+v", out)
	}
}
