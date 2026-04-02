package openaicompat

import (
	"io"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestStreamResponseEmitsToolDoneStopReasonAndCumulativeUsage(t *testing.T) {
	provider := &Provider{}
	body := strings.Join([]string{
		"data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"id\":\"call-1\",\"type\":\"function\",\"function\":{\"name\":\"fake\",\"arguments\":\"{\\\"x\\\":1}\"}}]},\"finish_reason\":null}]}",
		"data: {\"choices\":[{\"delta\":{},\"finish_reason\":\"tool_calls\"}],\"usage\":{\"prompt_tokens\":9,\"completion_tokens\":4}}",
		"data: [DONE]",
	}, "\n") + "\n"

	ch := make(chan types.StreamEvent, 16)
	go provider.streamResponse(io.NopCloser(strings.NewReader(body)), ch)

	var sawToolDone, sawUsage, sawMessageDone bool
	for ev := range ch {
		if ev.Type == types.StreamToolUseDone && ev.ToolUseID == "call-1" {
			sawToolDone = true
		}
		if ev.Usage != nil {
			sawUsage = true
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
		t.Fatalf("expected StreamToolUseDone for call-1")
	}
	if !sawUsage {
		t.Fatalf("expected usage event")
	}
	if !sawMessageDone {
		t.Fatalf("expected StreamMessageDone")
	}
}

func TestMapFinishReasonExpandedTaxonomy(t *testing.T) {
	tests := []struct {
		in   string
		want types.StopReason
	}{
		{in: "content_filter", want: types.StopContentFilter},
		{in: "context_length_exceeded", want: types.StopContextLimit},
		{in: "refusal", want: types.StopRefusal},
	}

	for _, tt := range tests {
		if got := mapFinishReason(tt.in); got != tt.want {
			t.Fatalf("mapFinishReason(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestStreamResponseTracksToolIDsAcrossIndexOnlyDeltas(t *testing.T) {
	provider := &Provider{}
	body := strings.Join([]string{
		"data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"id\":\"call-1\",\"type\":\"function\",\"function\":{\"name\":\"fake\",\"arguments\":\"{\"}}]},\"finish_reason\":null}]}",
		"data: {\"choices\":[{\"delta\":{\"tool_calls\":[{\"index\":0,\"type\":\"function\",\"function\":{\"arguments\":\"\\\"x\\\":1}\"}}]},\"finish_reason\":\"tool_calls\"}],\"usage\":{\"prompt_tokens\":10,\"completion_tokens\":2}}",
		"data: [DONE]",
	}, "\n") + "\n"

	ch := make(chan types.StreamEvent, 16)
	go provider.streamResponse(io.NopCloser(strings.NewReader(body)), ch)

	var sawDeltaWithID, sawDoneWithID bool
	for ev := range ch {
		if ev.Type == types.StreamToolUseDelta && ev.ToolUseID == "call-1" && ev.ToolName == "fake" {
			sawDeltaWithID = true
		}
		if ev.Type == types.StreamToolUseDone && ev.ToolUseID == "call-1" && ev.ToolName == "fake" {
			sawDoneWithID = true
		}
	}

	if !sawDeltaWithID {
		t.Fatalf("expected tool delta with resolved id/name")
	}
	if !sawDoneWithID {
		t.Fatalf("expected tool done with resolved id/name")
	}
}
