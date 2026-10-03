package mcp

import "testing"

func TestNormalizeServerStatusFieldsWave6(t *testing.T) {
	st := &ServerStatus{Authenticated: true, AuthStatus: AuthStatusUnauthenticated, ConnectionState: ServerConnectionNeedsAuth}
	normalizeServerStatusFields(st)
	if st.AuthStatus != AuthStatusAuthenticated || st.ConnectionState != ServerConnectionConnected {
		t.Fatalf("unexpected normalized status: %+v", st)
	}
}
