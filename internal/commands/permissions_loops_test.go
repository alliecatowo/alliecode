package commands

import (
	"context"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/permissions"
)

func TestPermissionsLoopsSubcommand(t *testing.T) {
	cmd := NewPermissionsCommand()
	state := &RuntimeState{PermissionMode: permissions.ModeBypass}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"loops"}})
	if err != nil {
		t.Fatalf("permissions loops failed: %v", err)
	}
	if !strings.Contains(res.Message, "PERMISSIONS_LOOPS") || !strings.Contains(res.Message, "loop.1.area=") {
		t.Fatalf("unexpected permissions loops output: %q", res.Message)
	}
}
