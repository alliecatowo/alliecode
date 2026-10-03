package tui

import (
	"strings"
	"testing"
)

func TestPermissionViewIncludesQueueAndRecentSections(t *testing.T) {
	model := NewPermission("bash", "execute tests")
	model.SetQueueIndex(1, 3)
	model.SetQueuePreview([]string{"read (turn 2) [PENDING]"})
	model.SetQueueStack([]string{"[active] bash [PENDING] turn 1 :: execute tests"})
	model.SetRecentDecisions([]string{"bash -> deny (turn 0)"})
	view := stripANSIForTest(model.View())
	for _, needle := range []string{"next in queue:", "stack:", "recent decisions:"} {
		if !strings.Contains(view, needle) {
			t.Fatalf("expected permission view to include %q, got %q", needle, view)
		}
	}
}
