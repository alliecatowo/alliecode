package commands

import (
	"context"
	"strings"
	"testing"
)

func TestSandboxRepair(t *testing.T) {
	cmd := NewSandboxCommand()
	state := &RuntimeState{SandboxMode: "danger-full-access"}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "sandbox", Args: []string{"repair"}})
	if err != nil || !strings.Contains(res.Message, "SANDBOX_REPAIR") {
		t.Fatalf("sandbox repair failed: %v %q", err, res.Message)
	}
}
