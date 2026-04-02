package commands

import (
	"context"
	"strings"
	"testing"
)

func TestConfigPanelSubcommand(t *testing.T) {
	cmd := NewConfigCommand()
	res, err := cmd.Execute(context.Background(), Context{State: &RuntimeState{}}, Invocation{Name: "config", Args: []string{"panel"}})
	if err != nil {
		t.Fatalf("config panel failed: %v", err)
	}
	if !strings.Contains(res.Message, "CONFIG_PANEL") || !strings.Contains(res.Message, "opened=true") {
		t.Fatalf("unexpected panel output: %q", res.Message)
	}
}
