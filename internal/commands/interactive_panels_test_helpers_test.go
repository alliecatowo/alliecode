package commands

import "testing"

func requirePanel(t *testing.T, name string, state *RuntimeState) InteractivePanel {
	t.Helper()
	panel, ok := InteractivePanelForCommand(name, Context{State: state})
	if !ok {
		t.Fatalf("expected interactive panel for %q", name)
	}
	if panel.Command != name {
		t.Fatalf("panel.Command = %q, want %q", panel.Command, name)
	}
	if len(panel.Items) == 0 {
		t.Fatalf("expected non-empty panel items for %q", name)
	}
	return panel
}
