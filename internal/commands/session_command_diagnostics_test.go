package commands

import (
	"context"
	"strings"
	"testing"
)

func TestSessionDiagnosticsIncludesQuickFixAndCounters(t *testing.T) {
	cmd := NewSessionCommand()
	state := &RuntimeState{SessionMode: "client", SessionConnected: false, SessionDiagnosticsCount: 1}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"diagnostics"}})
	if err != nil {
		t.Fatalf("session diagnostics failed: %v", err)
	}
	if !strings.Contains(res.Message, "SESSION_DIAGNOSTICS") || !strings.Contains(res.Message, "quick_fix=") {
		t.Fatalf("unexpected session diagnostics output: %q", res.Message)
	}
}
