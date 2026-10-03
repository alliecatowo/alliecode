package mcp

import "testing"

func TestStatusSummaryCountsWave6(t *testing.T) {
	s := summarizeServerStatuses([]ServerStatus{{ServerName: "a", ConnectionState: ServerConnectionConnected, Authenticated: true}, {ServerName: "b", ConnectionState: ServerConnectionFailed}})
	if s.Total != 2 || s.Authenticated != 1 || s.Failed != 1 {
		t.Fatalf("unexpected summary: %+v", s)
	}
}
