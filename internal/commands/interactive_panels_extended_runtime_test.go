package commands

import "testing"

func TestSupportsInteractivePanelIncludesExtendedRuntimeCommands(t *testing.T) {
	for _, name := range []string{"bridge", "sandbox-toggle", "mock-limits", "onboarding", "remote-setup", "tasks", "env", "issue", "workflows", "proactive", "assistant", "share", "compact"} {
		if !SupportsInteractivePanel(name) {
			t.Fatalf("expected interactive panel support for %q", name)
		}
	}
}

func TestInteractivePanelForExtendedRuntimeCommandsBuildsItems(t *testing.T) {
	state := &RuntimeState{}
	ctx := Context{State: state}
	for _, name := range []string{"bridge", "sandbox-toggle", "mock-limits", "onboarding", "remote-setup", "tasks", "env", "issue", "workflows", "proactive", "assistant", "share", "compact"} {
		panel, ok := InteractivePanelForCommand(name, ctx)
		if !ok {
			t.Fatalf("expected panel for %q", name)
		}
		if panel.Command == "" || len(panel.Items) == 0 {
			t.Fatalf("expected panel items for %q, got %+v", name, panel)
		}
	}
}
