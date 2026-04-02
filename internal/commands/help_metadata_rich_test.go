package commands

import (
	"context"
	"strings"
	"testing"
)

func TestHelpIncludesGroupHelpAndShortcuts(t *testing.T) {
	r := DefaultRegistry()
	cmd := NewHelpCommand(r)
	res, err := cmd.Execute(context.Background(), Context{}, Invocation{Name: "help", Args: []string{"provider"}})
	if err != nil {
		t.Fatalf("help failed: %v", err)
	}
	if !strings.Contains(res.Message, "entry.1.group=") || !strings.Contains(res.Message, "entry.1.help_hint=") || !strings.Contains(res.Message, "entry.1.shortcut_count=") {
		t.Fatalf("missing rich metadata fields: %q", res.Message)
	}
}
