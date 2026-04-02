package anthropic

import (
	"io"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStreamResponseEmitsToolDoneStopReasonAndCumulativeUsage(t *testing.T) {
	provider := &Provider{}
	body := strings.Join([]string{
		"data: {\"type\":\"message_start\"}",
		"data: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"tool_use\",\"id\":\"tool-1\",\"name\":\"fake\"}}",
		"data: {\"type\":\"content_block_delta\",\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":\"{\\\"x\\\":1}\"}}",
		"data: {\"type\":\"content_block_stop\",\"index\":0}",
		"data: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"tool_use\"},\"usage\":{\"output_tokens\":7}}",
		"data: {\"type\":\"message_stop\"}",
	}, "\n") + "\n"

	ch := make(chan types.StreamEvent, 16)
	go provider.streamResponse(io.NopCloser(strings.NewReader(body)), ch)

	var sawToolDone, sawMessageDone bool
	for ev := range ch {
		if ev.Type == types.StreamToolUseDone {
			sawToolDone = true
		}
		if ev.Usage != nil {
			if !ev.UsageCumulative {
				t.Fatalf("usage should be cumulative")
			}
			if ev.StopReason != types.StopToolUse {
				t.Fatalf("usage stop reason = %q, want %q", ev.StopReason, types.StopToolUse)
			}
		}
		if ev.Type == types.StreamMessageDone {
			sawMessageDone = true
			if ev.StopReason != types.StopToolUse {
				t.Fatalf("message done stop reason = %q, want %q", ev.StopReason, types.StopToolUse)
			}
		}
	}

	if !sawToolDone {
		t.Fatalf("expected StreamToolUseDone")
	}
	if !sawMessageDone {
		t.Fatalf("expected StreamMessageDone")
	}
}

func TestMapStopReasonExpandedTaxonomy(t *testing.T) {
	tests := []struct {
		in   string
		want types.StopReason
	}{
		{in: "model_context_window_exceeded", want: types.StopContextLimit},
		{in: "refusal", want: types.StopRefusal},
		{in: "end_turn", want: types.StopEndTurn},
	}

	for _, tt := range tests {
		if got := mapStopReason(tt.in); got != tt.want {
			t.Fatalf("mapStopReason(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestStreamResponseCarriesToolIdentityOnDeltaAndDone(t *testing.T) {
	provider := &Provider{}
	body := strings.Join([]string{
		"data: {\"type\":\"message_start\"}",
		"data: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"tool_use\",\"id\":\"tool-1\",\"name\":\"fake\"}}",
		"data: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"input_json_delta\",\"partial_json\":\"{\\\"x\\\":1}\"}}",
		"data: {\"type\":\"content_block_stop\",\"index\":0}",
		"data: {\"type\":\"message_stop\"}",
	}, "\n") + "\n"

	ch := make(chan types.StreamEvent, 16)
	go provider.streamResponse(io.NopCloser(strings.NewReader(body)), ch)

	var sawDeltaWithID, sawDoneWithID bool
	for ev := range ch {
		if ev.Type == types.StreamToolUseDelta && ev.ToolUseID == "tool-1" && ev.ToolName == "fake" {
			sawDeltaWithID = true
		}
		if ev.Type == types.StreamToolUseDone && ev.ToolUseID == "tool-1" && ev.ToolName == "fake" {
			sawDoneWithID = true
		}
	}

	if !sawDeltaWithID {
		t.Fatalf("expected tool delta with tool identity")
	}
	if !sawDoneWithID {
		t.Fatalf("expected tool done with tool identity")
	}
}
