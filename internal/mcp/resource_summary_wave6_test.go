package mcp

import "testing"

func TestResourceSummaryCountsWave6(t *testing.T) {
	s := summarizeResources([]Resource{{ServerName: "a", URI: "u1", MIMEType: "text/plain"}, {ServerName: "a", URI: "u2"}})
	if s.Total != 2 || s.ByServer["a"] != 2 || s.MissingMIME != 1 {
		t.Fatalf("unexpected resource summary: %+v", s)
	}
}
