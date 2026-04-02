package commands

import (
	"context"
	"strings"
	"testing"
)

func TestPluginDoctor(t *testing.T) {
	cmd := NewPluginCommand()
	state := &RuntimeState{}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "plugin", Args: []string{"doctor"}})
	if err != nil {
		t.Fatalf("plugin doctor failed: %v", err)
	}
	if !strings.Contains(res.Message, "PLUGIN_DOCTOR") {
		t.Fatalf("unexpected plugin doctor output: %q", res.Message)
	}
}
