package commands

import (
	"context"
	"strings"
	"testing"
)

func TestConfigRepairSupportsProfiles(t *testing.T) {
	cmd := NewConfigCommand()
	state := &RuntimeState{ConfigValues: map[string]string{}}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "config", Args: []string{"repair", "strict"}})
	if err != nil {
		t.Fatalf("config strict repair failed: %v", err)
	}
	if !strings.Contains(res.Message, "profile=strict") || state.ConfigLastRepairProfile != "strict" {
		t.Fatalf("unexpected strict profile behavior: %q profile=%q", res.Message, state.ConfigLastRepairProfile)
	}
}
