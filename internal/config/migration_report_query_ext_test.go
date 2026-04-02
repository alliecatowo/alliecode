package config

import "testing"

func TestQueryMigrationReportsStrictHydrationFilter(t *testing.T) {
	strict := true
	out := QueryMigrationReports([]MigrationReport{
		{HydrationMode: "strict", StrictHydration: true, DefaultProvider: "openai", MigrationVersion: 2, AppliedPatchCount: 1},
		{HydrationMode: "compat", StrictHydration: false, DefaultProvider: "openai", MigrationVersion: 2, AppliedPatchCount: 1},
	}, MigrationReportQuery{HydrationMode: "strict", Provider: "openai", StrictHydration: &strict})
	if len(out) != 1 || !out[0].StrictHydration {
		t.Fatalf("unexpected migration report query output: %+v", out)
	}
}
