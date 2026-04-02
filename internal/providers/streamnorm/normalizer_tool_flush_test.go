package streamnorm

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestNormalizeMessageDoneFlushesOpenTool(t *testing.T) {
	n := New("openai", "gpt-4o")
	_ = n.Normalize(types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: "u42", ToolName: "bash"})
	out := n.Normalize(types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn})

	found := false
	for _, ev := range out {
		if ev.Type == types.StreamToolBoundary && ev.Boundary == types.StreamBoundaryToolEnd {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected tool end boundary on message_done, got %+v", out)
	}
}
