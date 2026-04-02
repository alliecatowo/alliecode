package state

import "testing"

func TestBuildRuntimeHydrationSummaries(t *testing.T) {
	states := []RuntimeState{
		{SessionID: "s1", ProjectPath: "/repo/a", HydratedFromDisk: true, HydratedAtUnix: 100, StartupCount: 2},
		{SessionID: "s2", ProjectPath: "/repo/b", HydratedFromEnvironment: true, HydratedAtUnix: 200, StartupCount: 1, LastError: "boom"},
	}
	out := BuildRuntimeHydrationSummaries(states, RuntimeHydrationQuery{HydrationMode: "env", WithErrorsOnly: true, Limit: 5})
	if len(out) != 1 || out[0].SessionID != "s2" {
		t.Fatalf("unexpected hydration summary output: %+v", out)
	}
}

func TestBuildRuntimeHydrationCompatibility(t *testing.T) {
	states := []RuntimeState{{SessionID: "s1", HydratedFromDisk: true}}
	out := BuildRuntimeHydrationCompatibility(states, "env", StartupHydrationCompatibilityQuery{ConfiguredMode: "env", EffectiveMode: "disk", Limit: 10})
	if len(out) != 1 || !out[0].RequiresCompat {
		t.Fatalf("unexpected compatibility: %+v", out)
	}
}
