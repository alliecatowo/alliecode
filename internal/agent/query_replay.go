package agent

import "github.com/alliecatowo/alliecode/internal/types"

// QueryReplay returns a typed replay view for the given query.
func (a *Agent) QueryReplay(query types.AgentReplayQuery) types.AgentReplayView {
	events := a.ReplayEvents(query.FromSequence, query.ToSequence)
	filtered, truncated := filterReplayEvents(events, query)
	stats := buildReplayStats(events, filtered, query, truncated)
	window := replayWindow(events)
	dependencyGraph := buildContextDependencyGraph(filtered)

	return types.AgentReplayView{
		Query:           query,
		Window:          window,
		Cursor:          a.ReplayCursor(),
		Stats:           stats,
		DependencyGraph: dependencyGraph,
		Events:          filtered,
	}
}

// ReplayEventsForLabel returns replay checkpoint events with matching label.
func (a *Agent) ReplayEventsForLabel(label string, limit int) []types.AgentEvent {
	query := types.AgentReplayQuery{ReplayLabel: label, Limit: limit}
	return a.QueryReplay(query).Events
}

// ReplayEventsForStopReason returns stop events for matching reasons.
func (a *Agent) ReplayEventsForStopReason(reason types.AgentStopReason, limit int) []types.AgentEvent {
	query := types.AgentReplayQuery{StopReasons: []types.AgentStopReason{reason}, Limit: limit}
	return a.QueryReplay(query).Events
}

// ReplayEventsWithContinuity returns replay events preserving strict continuity.
func (a *Agent) ReplayEventsWithContinuity(query types.AgentReplayQuery) []types.AgentEvent {
	query.Continuity = types.AgentReplayContinuityStrict
	return a.QueryReplay(query).Events
}

// ReplayRewoundTo returns replay view clipped to the target sequence.
func (a *Agent) ReplayRewoundTo(targetSeq uint64, query types.AgentReplayQuery) types.AgentReplayView {
	query.RewindTo = targetSeq
	return a.QueryReplay(query)
}

// ReplayEventsByType returns replay events filtered by type.
func (a *Agent) ReplayEventsByType(eventType types.AgentEventType, limit int) []types.AgentEvent {
	query := types.AgentReplayQuery{Types: []types.AgentEventType{eventType}, Limit: limit}
	return a.QueryReplay(query).Events
}

// ReplayEventsForTurn returns replay events for a turn.
func (a *Agent) ReplayEventsForTurn(turn int, limit int) []types.AgentEvent {
	query := types.AgentReplayQuery{Turn: turn, Limit: limit}
	return a.QueryReplay(query).Events
}

// ReplayEventsForTurnIndex returns replay events for a turn index.
func (a *Agent) ReplayEventsForTurnIndex(turnIndex int, limit int) []types.AgentEvent {
	query := types.AgentReplayQuery{TurnIndex: turnIndex, Limit: limit}
	return a.QueryReplay(query).Events
}

// ReplayEventsByClass returns replay events filtered by event class.
func (a *Agent) ReplayEventsByClass(class types.AgentEventClass, limit int) []types.AgentEvent {
	query := types.AgentReplayQuery{Classes: []types.AgentEventClass{class}, Limit: limit}
	return a.QueryReplay(query).Events
}

// ReplayEventsByToolID returns replay events filtered by tool use ID.
func (a *Agent) ReplayEventsByToolID(toolID string, limit int) []types.AgentEvent {
	query := types.AgentReplayQuery{ToolID: toolID, Limit: limit}
	return a.QueryReplay(query).Events
}

func replayWindow(events []types.AgentEvent) types.AgentReplayWindow {
	if len(events) == 0 {
		return types.AgentReplayWindow{}
	}
	return types.AgentReplayWindow{
		Start: events[0].Sequence,
		End:   events[len(events)-1].Sequence,
		Total: len(events),
	}
}
