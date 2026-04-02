package state

import "testing"

func TestBuildStartupHydrationStatesRequireMigration(t *testing.T) {
	items := BuildStartupHydrationStates(
		[]RuntimeState{{SessionID: "s1", ProjectPath: "/repo/a", HydratedFromDisk: true}},
		[]SettingsCacheMetadata{{LastProjectDir: "/repo/a", LastScope: "layered"}},
		[]ConfigDiagnosticsState{{LastProjectPath: "/repo/a", LastMigrationVersion: "5"}},
		StartupHydrationQuery{RequireMigration: true, RequireCache: true, Limit: 5},
	)
	if len(items) != 1 || items[0].MigrationVersion != "5" || items[0].CacheScope != "layered" {
		t.Fatalf("unexpected startup hydration items: %+v", items)
	}
}
