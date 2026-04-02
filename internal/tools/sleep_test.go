package tools

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestSleepToolExecute(t *testing.T) {
	in, _ := json.Marshal(map[string]any{"duration_ms": 1})
	res, err := (&SleepTool{}).Execute(context.Background(), in, types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error: %s", res.Content)
	}
	if res.Content != `{"status":"completed","requested_ms":1,"slept_ms":1,"bounded":false}` {
		t.Fatalf("unexpected deterministic payload: %s", res.Content)
	}
}

func TestSleepToolBoundsDurationLow(t *testing.T) {
	in, _ := json.Marshal(map[string]any{"duration_ms": 0})
	res, err := (&SleepTool{}).Execute(context.Background(), in, types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result: %s", res.Content)
	}
	if res.Content != `{"status":"completed","requested_ms":0,"slept_ms":1,"bounded":true}` {
		t.Fatalf("unexpected deterministic bounded payload: %s", res.Content)
	}
}

func TestSleepToolBoundsDurationHigh(t *testing.T) {
	in, _ := json.Marshal(map[string]any{"duration_ms": 600001})
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	res, err := (&SleepTool{}).Execute(ctx, in, types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected cancellation when bounded sleep exceeds context timeout")
	}
	if res.Content != `{"status":"cancelled","requested_ms":600001,"slept_ms":600000,"bounded":true}` {
		t.Fatalf("unexpected deterministic cancellation payload: %s", res.Content)
	}
}
