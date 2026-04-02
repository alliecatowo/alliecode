package settings

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/state"
)

func TestManagerBuildStartupHydrationStates(t *testing.T) {
	m := NewManager(0)
	items := m.BuildStartupHydrationStates(
		[]state.RuntimeState{{SessionID: "s1", ProjectPath: "/repo/a", HydratedFromEnvironment: true}},
		[]state.SettingsCacheMetadata{{LastProjectDir: "/repo/a", LastScope: "layered"}},
		[]state.ConfigDiagnosticsState{{LastProjectPath: "/repo/a", LastMigrationVersion: "1"}},
		state.StartupHydrationQuery{RequireCache: true, Limit: 5},
	)
	if len(items) != 1 || items[0].SessionID != "s1" {
		t.Fatalf("unexpected hydration items: %+v", items)
	}
}
