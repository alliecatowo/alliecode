package tools

import (
	"context"
	"testing"
)

func TestDefaultRegistryIncludesNewUtilityTools(t *testing.T) {
	r := DefaultRegistry()

	for _, name := range []string{
		"powershell",
		"enter_plan_mode",
		"exit_plan_mode",
		"repl",
		"synthetic_output",
		"ls",
		"task_list",
		"memory",
		"AskUserQuestion",
		"Sleep",
		"task_get",
		"task_create",
		"task_update",
		"task_stop",
		"task_output",
		"tool_search",
		"skill",
		"config",
		"lsp",
		"worktree_enter",
		"worktree_exit",
		"send_message",
		"team_create",
		"team_list",
		"team_status",
		"team_update",
		"team_delete",
		"cron_create",
		"cron_delete",
		"cron_list",
		"remote_trigger",
		"brief",
	} {
		if _, ok := r.Get(name); !ok {
			t.Fatalf("expected tool %q to be registered", name)
		}
	}
}

func TestDefaultRegistryWithMCPRegistersResourceWrappers(t *testing.T) {
	r, mgr, err := DefaultRegistryWithMCP(context.Background(), nil)
	if err != nil {
		t.Fatalf("DefaultRegistryWithMCP error: %v", err)
	}
	if mgr != nil {
		t.Fatalf("expected nil mcp manager for empty config")
	}

	for _, name := range []string{"mcp_resource_list", "mcp_resource_read", "mcp_tool_invoke", "mcp_auth_local", "mcp_auth_status"} {
		if _, ok := r.Get(name); ok {
			t.Fatalf("did not expect MCP wrapper %q without MCP config", name)
		}
	}
}
