package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/mcp"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestMCPListSuccessIncludesContractWave6(t *testing.T) {
	fake := &fakeRegistryMCPManager{
		resources: []mcp.Resource{{ServerName: "srv", URI: "mem://x", Name: "n"}},
		transport: map[string]mcp.TransportType{"srv": mcp.TransportStdio},
	}
	res, err := (&MCPResourceListTool{manager: fake}).Execute(context.Background(), []byte(`{}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("unexpected failure: err=%v content=%q", err, res.Content)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["contract"] == nil {
		t.Fatalf("expected contract metadata")
	}
}
