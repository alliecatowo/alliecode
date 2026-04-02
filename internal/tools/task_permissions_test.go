package tools

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskFamilyPermissionBridge(t *testing.T) {
	ctx := types.ToolContext{IsNonInteractive: true}
	if got := (&TaskListTool{}).CheckPermissions([]byte(`{}`), ctx); got != types.PermissionAllowed {
		t.Fatalf("task_list permission = %v, want allow", got)
	}
	if got := (&TaskUpdateTool{}).CheckPermissions([]byte(`{"task_id":"x","status":"completed"}`), ctx); got != types.PermissionAsk {
		t.Fatalf("task_update permission = %v, want ask", got)
	}
}
