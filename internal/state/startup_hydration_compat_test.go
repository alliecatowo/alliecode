package state

import "testing"

func TestNormalizeHydrationMode(t *testing.T) {
	if got := NormalizeHydrationMode("legacy"); got != "disk" {
		t.Fatalf("NormalizeHydrationMode(legacy) = %q", got)
	}
	if got := NormalizeHydrationMode(""); got != "compat" {
		t.Fatalf("NormalizeHydrationMode(empty) = %q", got)
	}
}

func TestBuildStartupHydrationCompatibility(t *testing.T) {
	states := []RuntimeState{{SessionID: "s1", ProjectPath: "/repo/a", HydratedFromEnvironment: true, RuntimeSurface: "cli"}}
	out := BuildStartupHydrationCompatibility(states, "disk", StartupHydrationCompatibilityQuery{ConfiguredMode: "disk", EffectiveMode: "env", RequireCompat: boolPtr(true), Limit: 5})
	if len(out) != 1 || !out[0].RequiresCompat || out[0].EffectiveMode != "env" {
		t.Fatalf("unexpected compatibility output: %+v", out)
	}
}
