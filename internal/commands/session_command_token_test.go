package commands

import (
	"context"
	"strings"
	"testing"
)

func TestSessionTokenSubcommands(t *testing.T) {
	cmd := NewSessionCommand()
	state := &RuntimeState{}
	setRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"token", "set", "abc123token"}})
	if err != nil || !strings.Contains(setRes.Message, "SESSION_TOKEN_SET") {
		t.Fatalf("set token failed: %v %q", err, setRes.Message)
	}
	showRes, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"token"}})
	if err != nil || !strings.Contains(showRes.Message, "set=true") {
		t.Fatalf("show token failed: %v %q", err, showRes.Message)
	}
}
