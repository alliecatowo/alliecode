package streamnorm

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestNormalizeAddsRequestBoundaryAndUsageEvents(t *testing.T) {
	n := New("anthropic", "claude-sonnet-4")
	out := n.Normalize(types.StreamEvent{Type: types.StreamContentDelta, Delta: "hi", Usage: &types.Usage{InputTokens: 2, OutputTokens: 1}})
	if len(out) < 4 {
		t.Fatalf("expected >= 4 events, got %d", len(out))
	}
	if out[0].Type != types.StreamRequestStart {
		t.Fatalf("first event = %q, want request_start", out[0].Type)
	}
	if out[2].Type != types.StreamUsageDelta {
		t.Fatalf("third event = %q, want usage_delta", out[2].Type)
	}
	if out[3].Type != types.StreamUsageTotal || !out[3].UsageCumulative {
		t.Fatalf("fourth event should be cumulative usage total")
	}
}

func TestNormalizeToolBoundaries(t *testing.T) {
	n := New("openai", "gpt-4o")
	start := n.Normalize(types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: "tool-1", ToolName: "bash"})
	if start[len(start)-1].Type != types.StreamToolBoundary || start[len(start)-1].Boundary != types.StreamBoundaryToolBegin {
		t.Fatalf("expected tool begin boundary")
	}
	done := n.Normalize(types.StreamEvent{Type: types.StreamToolUseDone, ToolUseID: "tool-1", ToolName: "bash"})
	if done[len(done)-1].Type != types.StreamToolBoundary || done[len(done)-1].Boundary != types.StreamBoundaryToolEnd {
		t.Fatalf("expected tool end boundary")
	}
}

func TestNormalizeCumulativeUsageDelta(t *testing.T) {
	n := New("gemini", "gemini-2.5-pro")
	_ = n.Normalize(types.StreamEvent{Type: types.StreamContentDelta, Usage: &types.Usage{InputTokens: 5, OutputTokens: 2}, UsageCumulative: true})
	out := n.Normalize(types.StreamEvent{Type: types.StreamContentDelta, Usage: &types.Usage{InputTokens: 8, OutputTokens: 3}, UsageCumulative: true})
	var delta *types.Usage
	for _, ev := range out {
		if ev.Type == types.StreamUsageDelta {
			delta = ev.Usage
			break
		}
	}
	if delta == nil {
		t.Fatalf("expected usage delta event")
	}
	if delta.InputTokens != 3 || delta.OutputTokens != 1 {
		t.Fatalf("delta = %+v, want input=3 output=1", *delta)
	}
}
