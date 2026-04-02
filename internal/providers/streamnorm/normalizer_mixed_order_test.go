package streamnorm

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestNormalizeClosesToolBeforeContentResumes(t *testing.T) {
	n := New("anthropic", "claude-sonnet")
	_ = n.Normalize(types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: "t1", ToolName: "edit"})
	out := n.Normalize(types.StreamEvent{Type: types.StreamContentDelta, Delta: "hello"})

	foundToolEnd := false
	foundContent := false
	for _, ev := range out {
		if ev.Type == types.StreamToolBoundary && ev.Boundary == types.StreamBoundaryToolEnd {
			foundToolEnd = true
		}
		if ev.Type == types.StreamContentDelta {
			foundContent = true
		}
	}
	if !foundToolEnd || !foundContent {
		t.Fatalf("expected tool_end boundary and content delta, got %+v", out)
	}
}
