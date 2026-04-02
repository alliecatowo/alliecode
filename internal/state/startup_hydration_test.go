package state

import "testing"

func TestBuildStartupHydrationStates(t *testing.T) {
	runtimeStates := []RuntimeState{{SessionID: "s1", ProjectPath: "/repo/a", HydratedFromDisk: true, HydratedAtUnix: 5, StartupCount: 3}}
	cache := []SettingsCacheMetadata{{LastProjectDir: "/repo/a", LastScope: "layered", LastKnownLayeringMode: "layered", LastPersistedAtUnixNano: 99, Fresh: true, InvalidateCount: 1}}
	diags := []ConfigDiagnosticsState{{LastProjectPath: "/repo/a", LastValidationPassed: true, LastValidationErrorCount: 1, LastValidationWarningCount: 2, LastMigrationVersion: "2", MigrationAppliedCount: 2}}
	out := BuildStartupHydrationStates(runtimeStates, cache, diags, StartupHydrationQuery{Hydrated: boolPtr(true), Validation: boolPtr(true), RequireCache: true, RequireMigration: true, Limit: 5})
	if len(out) != 1 || out[0].MigrationVersion != "2" || !out[0].CacheFresh || out[0].ValidationErrors != 1 || out[0].CacheLayeringMode != "layered" {
		t.Fatalf("unexpected startup hydration states: %+v", out)
	}
}

func boolPtr(v bool) *bool { return &v }
