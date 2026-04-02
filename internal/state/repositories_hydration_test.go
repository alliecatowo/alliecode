package state

import "testing"

func TestRepositoriesBuildStartupHydrationSummary(t *testing.T) {
	paths, err := ResolvePaths(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatalf("ResolvePaths() error = %v", err)
	}
	repos := NewRepositories(paths)
	out := repos.BuildStartupHydrationSummary(
		[]RuntimeState{{SessionID: "s1", ProjectPath: "/repo/a", HydratedFromEnvironment: true, StartupCount: 1}},
		[]SettingsCacheMetadata{{LastProjectDir: "/repo/a", Fresh: true}},
		[]ConfigDiagnosticsState{{LastProjectPath: "/repo/a", LastValidationPassed: true}},
		StartupHydrationQuery{Limit: 1},
	)
	if len(out) != 1 || out[0].SessionID != "s1" {
		t.Fatalf("unexpected startup hydration summary: %+v", out)
	}
}

func TestRepositoriesBuildStartupHydrationSummaryPage(t *testing.T) {
	paths, err := ResolvePaths(t.TempDir(), t.TempDir())
	if err != nil {
		t.Fatalf("ResolvePaths() error = %v", err)
	}
	repos := NewRepositories(paths)
	items, total := repos.BuildStartupHydrationSummaryPage(
		[]RuntimeState{{SessionID: "s1", ProjectPath: "/repo/a", HydratedFromEnvironment: true, StartupCount: 1}, {SessionID: "s2", ProjectPath: "/repo/b", HydratedFromDisk: true, StartupCount: 2}},
		nil,
		nil,
		StartupHydrationQuery{Limit: 10},
		1,
		1,
	)
	if total != 2 || len(items) != 1 {
		t.Fatalf("unexpected startup hydration page: total=%d items=%+v", total, items)
	}
}
