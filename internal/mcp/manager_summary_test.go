package mcp

import "testing"

func TestManagerSummaryHelpers(t *testing.T) {
	m := &Manager{
		connectionState: map[string]ServerConnectionState{"a": ServerConnectionConnected},
		authenticated:   map[string]bool{"a": true},
		authStatus:      map[string]AuthStatus{"a": AuthStatusAuthenticated},
		transportType:   map[string]TransportType{"a": TransportStdio},
		resourceList:    []Resource{{ServerName: "a", URI: "mem://1", MIMEType: "text/plain"}},
	}
	ss := m.ServerStatusSummary()
	if ss.Total != 1 || ss.Authenticated != 1 {
		t.Fatalf("unexpected server summary: %+v", ss)
	}
	rs := m.ResourceSummary()
	if rs.Total != 1 || rs.ByServer["a"] != 1 {
		t.Fatalf("unexpected resource summary: %+v", rs)
	}
}
