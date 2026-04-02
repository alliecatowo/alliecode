package commands

import (
	"context"
	"strings"
	"testing"
)

func TestConfigRepair(t *testing.T) {
	cmd := NewConfigCommand()
	state := &RuntimeState{}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "config", Args: []string{"repair"}})
	if err != nil || !strings.Contains(res.Message, "CONFIG_REPAIR") {
		t.Fatalf("config repair failed: %v %q", err, res.Message)
	}
}
