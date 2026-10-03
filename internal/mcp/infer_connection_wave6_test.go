package mcp

import "testing"

func TestInferServerConnectionStateWave6(t *testing.T) {
	if got := inferServerConnectionState(false, AuthStatusUnauthenticated); got != ServerConnectionNeedsAuth {
		t.Fatalf("unexpected state: %q", got)
	}
	if got := inferServerConnectionState(true, AuthStatusAuthenticated); got != ServerConnectionConnected {
		t.Fatalf("unexpected state: %q", got)
	}
}
