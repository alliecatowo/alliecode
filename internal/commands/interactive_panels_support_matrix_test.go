package commands

import "testing"

func TestSupportsInteractivePanelExpandedRuntimeFamilies(t *testing.T) {
	for _, name := range []string{"model", "provider", "permissions", "doctor", "status", "mcp", "plugin", "skills", "history", "session", "branch", "diff", "files", "memory", "theme", "output-style", "privacy-settings", "upgrade", "resume", "plan", "tasks", "env", "issue", "workflows", "proactive", "assistant", "share", "compact"} {
		if !SupportsInteractivePanel(name) {
			t.Fatalf("expected %q to be supported", name)
		}
	}
	for _, name := range []string{"help", "login", "logout", "not-a-command"} {
		if SupportsInteractivePanel(name) {
			t.Fatalf("expected %q to remain unsupported", name)
		}
	}
}
