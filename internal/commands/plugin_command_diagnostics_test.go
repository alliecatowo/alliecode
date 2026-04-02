package commands

import (
	"context"
	"strings"
	"testing"
)

func TestPluginDiagnosticsAndRepair(t *testing.T) {
	cmd := NewPluginCommand()
	state := &RuntimeState{PluginsInstalled: []string{"acme/foo"}, PluginReloadPending: true}
	diagRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "plugin", Args: []string{"diagnostics"}})
	if err != nil {
		t.Fatalf("plugin diagnostics failed: %v", err)
	}
	if !strings.Contains(diagRes.Message, "PLUGIN_DIAGNOSTICS") {
		t.Fatalf("unexpected plugin diagnostics: %q", diagRes.Message)
	}
	repairRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "plugin", Args: []string{"repair", "reload"}})
	if err != nil {
		t.Fatalf("plugin repair failed: %v", err)
	}
	if !strings.Contains(repairRes.Message, "PLUGIN_REPAIR") || state.PluginReloadPending {
		t.Fatalf("expected reload repair to clear pending reload: %q", repairRes.Message)
	}
}
