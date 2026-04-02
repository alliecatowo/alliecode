package agent

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestEnforceStrictReplayContinuityFillsSequenceGaps(t *testing.T) {
	all := []types.AgentEvent{{Sequence: 1}, {Sequence: 2}, {Sequence: 3}, {Sequence: 4}}
	filtered := []types.AgentEvent{{Sequence: 1}, {Sequence: 4}}
	out := enforceStrictReplayContinuity(all, filtered)
	if len(out) != 4 {
		t.Fatalf("strict continuity events = %d, want 4", len(out))
	}
}

func TestReplayContinuityTracksLastLabel(t *testing.T) {
	events := []types.AgentEvent{{Sequence: 2, ReplayLabel: "turn_begin"}, {Sequence: 4, ReplayLabel: "turn_end"}}
	cont := replayContinuity(events, types.AgentReplayContinuityStrict)
	if cont.SequenceGaps != 1 {
		t.Fatalf("sequence gaps = %d, want 1", cont.SequenceGaps)
	}
	if cont.LastReplayLabel != "turn_end" {
		t.Fatalf("last label = %q, want turn_end", cont.LastReplayLabel)
	}
}
