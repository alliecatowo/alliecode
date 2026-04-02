package agent

import "github.com/alliecatowo/alliecode/internal/types"

func buildReplayStats(allEvents, events []types.AgentEvent, query types.AgentReplayQuery, truncated bool) types.AgentReplayStats {
	if query.Continuity == "" {
		query.Continuity = types.AgentReplayContinuityNone
	}

	if len(events) == 0 {
		return types.AgentReplayStats{
			Continuity: replayContinuity(events, query.Continuity),
			Rewind:     replayRewindState(allEvents, query),
		}
	}

	if len(allEvents) == 0 {
		allEvents = events
	}

	typeCounts := map[types.AgentEventType]int{}
	classCounts := map[types.AgentEventClass]int{}
	turns := map[int]struct{}{}
	tools := map[string]struct{}{}
	hasStop := false
	for _, ev := range events {
		typeCounts[ev.Type]++
		classCounts[eventClassFor(ev.Type)]++
		if ev.Turn > 0 {
			turns[ev.Turn] = struct{}{}
		}
		if ev.ToolUseID != "" {
			tools[ev.ToolUseID] = struct{}{}
		}
		if ev.Type == types.AgentEventStop {
			hasStop = true
		}
	}

	list := make([]types.AgentEventTypeCount, 0, len(typeCounts))
	for t, c := range typeCounts {
		list = append(list, types.AgentEventTypeCount{Type: t, Count: c})
	}
	classList := make([]types.AgentEventClassCount, 0, len(classCounts))
	for class, count := range classCounts {
		classList = append(classList, types.AgentEventClassCount{Class: class, Count: count})
	}

	return types.AgentReplayStats{
		Matched:      len(events),
		Truncated:    truncated,
		TypeCounts:   list,
		ClassCounts:  classList,
		FirstSeq:     events[0].Sequence,
		LastSeq:      events[len(events)-1].Sequence,
		TurnsSeen:    len(turns),
		ToolsSeen:    len(tools),
		HasStopEvent: hasStop,
		Continuity:   replayContinuity(events, query.Continuity),
		Rewind:       replayRewindState(allEvents, query),
	}
}
