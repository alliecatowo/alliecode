package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/alliecatowo/alliecode/internal/mcp"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestMCPErrorPayloadIncludesContractWave6(t *testing.T) {
	fake := &fakeRegistryMCPManager{listErr: errors.New("boom")}
	res, err := (&MCPResourceListTool{manager: fake}).Execute(context.Background(), []byte(`{}`), types.ToolContext{})
	if err != nil || !res.IsError {
		t.Fatalf("expected error result: err=%v content=%q", err, res.Content)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["contract"] == nil || out["category"] == nil {
		t.Fatalf("expected contract and category fields: %+v", out)
	}
	_ = mcp.AuthStatusAuthenticated
}
