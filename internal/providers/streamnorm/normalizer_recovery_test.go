package streamnorm

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestNormalizeRecoversToolDeltaBeforeStart(t *testing.T) {
	n := New("openai", "gpt-4o")
	out := n.Normalize(types.StreamEvent{Type: types.StreamToolUseDelta, ToolUseID: "u1", ToolName: "bash", Delta: "{}"})
	if len(out) < 4 {
		t.Fatalf("expected recovery events, got %d", len(out))
	}
	if out[1].Type != types.StreamToolUseStart {
		t.Fatalf("expected synthesized tool_use_start, got %s", out[1].Type)
	}
	if out[2].Type != types.StreamToolBoundary || out[2].Boundary != types.StreamBoundaryToolBegin {
		t.Fatalf("expected tool begin boundary, got %+v", out[2])
	}
}
