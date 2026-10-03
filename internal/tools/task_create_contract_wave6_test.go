package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestTaskCreateContractMetadataWave6(t *testing.T) {
	res, err := (&TaskCreateTool{}).Execute(context.Background(), []byte(`{"subject":"s","description":"d"}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("task_create failed: err=%v content=%q", err, res.Content)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out["contract"] == nil {
		t.Fatalf("expected contract metadata")
	}
}
