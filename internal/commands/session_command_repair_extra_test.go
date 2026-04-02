package commands

import (
	"context"
	"strings"
	"testing"
)

func TestSessionRepairAutoCreatesTokenWhenMissing(t *testing.T) {
	cmd := NewSessionCommand()
	state := &RuntimeState{}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "session", Args: []string{"repair"}})
	if err != nil {
		t.Fatalf("session repair failed: %v", err)
	}
	if !strings.Contains(res.Message, "SESSION_REPAIR") || strings.TrimSpace(state.SessionToken) == "" {
		t.Fatalf("expected token to be repaired: %q token=%q", res.Message, state.SessionToken)
	}
}
