package commands

import (
	"context"
	"testing"
)

func TestMCPNoRedirectSubcommand(t *testing.T) {
	cmd := NewMCPCommand()
	res, err := cmd.Execute(context.Background(), Context{State: &RuntimeState{}}, Invocation{Name: "mcp", Args: []string{"no-redirect"}})
	if err != nil {
		t.Fatalf("mcp no-redirect failed: %v", err)
	}
	if res.Message != "MCP_SETTINGS\nredirect=false" {
		t.Fatalf("unexpected no-redirect output: %q", res.Message)
	}
}
