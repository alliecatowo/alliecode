package settings

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/state"
)

func TestManagerRuntimeHydrationCompatibility(t *testing.T) {
	manager := NewManager(0)
	states := []state.RuntimeState{{SessionID: "s1", ProjectPath: "/repo/a", HydratedFromEnvironment: true}}
	items := manager.RuntimeHydrationCompatibility(states, "disk", state.StartupHydrationCompatibilityQuery{ConfiguredMode: "disk", EffectiveMode: "env", Limit: 5})
	if len(items) != 1 || !items[0].RequiresCompat {
		t.Fatalf("unexpected compatibility items: %+v", items)
	}
}
