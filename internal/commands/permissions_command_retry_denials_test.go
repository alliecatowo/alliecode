package commands

import (
	"context"
	"strings"
	"testing"
)

func TestPermissionsRetryDenials(t *testing.T) {
	cmd := NewPermissionsCommand()
	state := &RuntimeState{PermissionDenials: []PermissionDenial{{ID: "1", Command: "git push", Reason: "network"}, {ID: "2", Command: "git push", Reason: "network"}, {ID: "3", Command: "go test ./...", Reason: "policy"}}}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "permissions", Args: []string{"retry-denials"}})
	if err != nil {
		t.Fatalf("permissions retry-denials failed: %v", err)
	}
	if !strings.Contains(res.Message, "PERMISSIONS_RETRY") || !strings.Contains(res.Message, "count=2") {
		t.Fatalf("unexpected retry-denials output: %q", res.Message)
	}
}
