package tools

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/mcp"
)

func TestSummarizeMCPResources(t *testing.T) {
	resources := []mcp.Resource{{ServerName: "a", URI: "u1", MIMEType: "text/plain"}, {ServerName: "a", URI: "u2"}, {ServerName: "b", URI: "u3", Name: "n", MIMEType: "application/json"}}
	statuses := []mcp.ServerStatus{{ServerName: "a", ConnectionState: mcp.ServerConnectionConnected}, {ServerName: "b", ConnectionState: mcp.ServerConnectionNeedsAuth}}
	s := summarizeMCPResources(resources, statuses)
	if s.UniqueServers != 2 || s.ByServer["a"] != 2 || s.MissingMIME != 1 || s.StatusOverview[string(mcp.ServerConnectionNeedsAuth)] != 1 {
		t.Fatalf("unexpected summary: %+v", s)
	}
}
