package streamnorm

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestNormalizeHandlesCumulativeUsageReset(t *testing.T) {
	n := New("openai", "o4-mini")
	_ = n.Normalize(types.StreamEvent{Type: types.StreamContentDelta, Usage: &types.Usage{InputTokens: 100, OutputTokens: 20}, UsageCumulative: true})
	out := n.Normalize(types.StreamEvent{Type: types.StreamContentDelta, Usage: &types.Usage{InputTokens: 3, OutputTokens: 1}, UsageCumulative: true})

	for _, ev := range out {
		if ev.Type == types.StreamUsageDelta {
			if ev.Usage.InputTokens != 3 || ev.Usage.OutputTokens != 1 {
				t.Fatalf("unexpected reset delta: %+v", *ev.Usage)
			}
			return
		}
	}
	t.Fatalf("expected usage delta event in %+v", out)
}
