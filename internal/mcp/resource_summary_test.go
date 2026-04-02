package mcp

import "testing"

func TestSummarizeResources(t *testing.T) {
	resources := []Resource{{ServerName: "a", URI: "u1", MIMEType: "text/plain"}, {ServerName: "a", URI: "u2"}, {ServerName: "b", URI: "u3", Name: "n", MIMEType: "application/json"}}
	s := summarizeResources(resources)
	if s.Total != 3 || s.ByServer["a"] != 2 || s.MissingMIME != 1 || s.UniqueServer != 2 {
		t.Fatalf("unexpected resource summary: %+v", s)
	}
}
