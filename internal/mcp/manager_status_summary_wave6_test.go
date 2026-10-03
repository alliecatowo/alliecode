package mcp

import "testing"

func TestManagerServerStatusSummaryWave6(t *testing.T) {
	m := &Manager{connectionState: map[string]ServerConnectionState{"a": ServerConnectionConnected}, authStatus: map[string]AuthStatus{"a": AuthStatusAuthenticated}, authenticated: map[string]bool{"a": true}, transportType: map[string]TransportType{"a": TransportStdio}}
	s := m.ServerStatusSummary()
	if s.Total != 1 || s.Authenticated != 1 {
		t.Fatalf("unexpected status summary: %+v", s)
	}
}
