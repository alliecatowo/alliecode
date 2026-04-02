package streamnorm

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestNormalizeDropsLateContentAfterTurnEnd(t *testing.T) {
	n := New("gemini", "gemini-2.5-pro")
	_ = n.Normalize(types.StreamEvent{Type: types.StreamMessageDone, StopReason: types.StopEndTurn})
	out := n.Normalize(types.StreamEvent{Type: types.StreamContentDelta, Delta: "late"})
	if len(out) != 0 {
		t.Fatalf("expected no late content after turn end, got %+v", out)
	}
}
