package tools

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestEvaluateToolPermissionByFamilyNonInteractiveMutationAsks(t *testing.T) {
	got := evaluateToolPermissionByFamily(toolFamilyTeam, "team_update", nil, types.ToolContext{IsNonInteractive: true}, types.PermissionAllowed)
	if got != types.PermissionAsk {
		t.Fatalf("expected ask, got %v", got)
	}
}
