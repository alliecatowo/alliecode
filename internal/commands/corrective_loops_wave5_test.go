package commands

import "testing"

func TestCorrectiveLoopsIncludeWorkflowFamilies(t *testing.T) {
	loops := correctiveLoopsForState(&RuntimeState{})
	areas := map[string]bool{}
	for _, loop := range loops {
		areas[loop.Area] = true
	}
	for _, area := range []string{"provider", "model", "permissions", "settings", "history", "session", "mcp"} {
		if !areas[area] {
			t.Fatalf("missing corrective loop area %q in %#v", area, loops)
		}
	}
}
