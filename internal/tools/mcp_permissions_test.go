package tools

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestMCPFamilyPermissionBridge(t *testing.T) {
	ctx := types.ToolContext{IsNonInteractive: true}
	if got := (&MCPAuthStatusTool{}).CheckPermissions([]byte(`{"server_name":"alpha"}`), ctx); got != types.PermissionAllowed {
		t.Fatalf("mcp_auth_status permission = %v, want allow", got)
	}
	if got := (&MCPToolInvokeTool{}).CheckPermissions([]byte(`{"server_name":"alpha","tool_name":"sum"}`), ctx); got != types.PermissionAsk {
		t.Fatalf("mcp_tool_invoke permission = %v, want ask", got)
	}
}
