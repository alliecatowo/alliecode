package mcp

import "testing"

func TestSummarizeServerStatuses(t *testing.T) {
	statuses := []ServerStatus{{ServerName: "a", ConnectionState: ServerConnectionConnected, Authenticated: true}, {ServerName: "b", ConnectionState: ServerConnectionNeedsAuth}}
	s := summarizeServerStatuses(statuses)
	if s.Total != 2 || s.Authenticated != 1 || s.NeedsAuth != 1 || s.ByConnection[string(ServerConnectionConnected)] != 1 {
		t.Fatalf("unexpected status summary: %+v", s)
	}
}
