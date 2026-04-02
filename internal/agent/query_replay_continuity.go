package agent

import "github.com/alliecatowo/alliecode/internal/types"

func enforceStrictReplayContinuity(allEvents, filtered []types.AgentEvent) []types.AgentEvent {
	if len(filtered) <= 1 {
		return filtered
	}
	minSeq := filtered[0].Sequence
	maxSeq := filtered[len(filtered)-1].Sequence
	for _, ev := range filtered[1:] {
		if ev.Sequence < minSeq {
			minSeq = ev.Sequence
		}
		if ev.Sequence > maxSeq {
			maxSeq = ev.Sequence
		}
	}
	out := make([]types.AgentEvent, 0, int(maxSeq-minSeq)+1)
	for _, ev := range allEvents {
		if ev.Sequence < minSeq || ev.Sequence > maxSeq {
			continue
		}
		out = append(out, ev)
	}
	return out
}

func replayContinuity(events []types.AgentEvent, mode types.AgentReplayContinuityMode) types.AgentReplayContinuity {
	out := types.AgentReplayContinuity{Mode: mode, Strict: mode == types.AgentReplayContinuityStrict}
	if len(events) == 0 {
		return out
	}
	lastSeq := events[0].Sequence
	for _, ev := range events {
		if ev.Sequence > lastSeq+1 {
			out.SequenceGaps += int(ev.Sequence-lastSeq) - 1
		}
		lastSeq = ev.Sequence
		if ev.ReplayLabel != "" {
			out.LastReplayLabel = ev.ReplayLabel
		}
	}
	if out.Strict {
		out.MissingParentRefs = missingDependencyParents(events)
	}
	return out
}

func replayCheckpoints(events []types.AgentEvent) []types.AgentRewindPoint {
	out := make([]types.AgentRewindPoint, 0)
	for _, ev := range events {
		if ev.Type != types.AgentEventReplayCheckpoint {
			continue
		}
		out = append(out, types.AgentRewindPoint{
			Sequence:    ev.Sequence,
			Turn:        ev.Turn,
			ReplayLabel: ev.ReplayLabel,
		})
	}
	return out
}

func replayRewindState(allEvents []types.AgentEvent, query types.AgentReplayQuery) types.AgentReplayRewindState {
	state := types.AgentReplayRewindState{Checkpoints: replayCheckpoints(allEvents)}
	if query.RewindTo == 0 {
		return state
	}
	state.Applied = true
	state.TargetSeq = query.RewindTo
	for _, cp := range state.Checkpoints {
		if cp.Sequence == query.RewindTo {
			state.Checkpointed = true
			state.TargetLabel = cp.ReplayLabel
			break
		}
	}
	return state
}
