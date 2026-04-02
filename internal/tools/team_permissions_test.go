package tools

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTeamFamilyPermissionBridge(t *testing.T) {
	ctx := types.ToolContext{IsNonInteractive: true}
	if got := (&TeamListTool{}).CheckPermissions([]byte(`{}`), ctx); got != types.PermissionAllowed {
		t.Fatalf("team_list permission = %v, want allow", got)
	}
	if got := (&TeamDeleteTool{}).CheckPermissions([]byte(`{"team_name":"alpha"}`), ctx); got != types.PermissionAsk {
		t.Fatalf("team_delete permission = %v, want ask", got)
	}
}
