package agent

import "github.com/alliecatowo/alliecode/internal/types"

func filterReplayEvents(events []types.AgentEvent, query types.AgentReplayQuery) ([]types.AgentEvent, bool) {
	typeSet := make(map[types.AgentEventType]struct{}, len(query.Types))
	for _, t := range query.Types {
		typeSet[t] = struct{}{}
	}
	classSet := make(map[types.AgentEventClass]struct{}, len(query.Classes))
	for _, class := range query.Classes {
		classSet[class] = struct{}{}
	}
	stopSet := make(map[types.AgentStopReason]struct{}, len(query.StopReasons))
	for _, reason := range query.StopReasons {
		stopSet[reason] = struct{}{}
	}

	turnFilter := query.Turn
	if turnFilter <= 0 && query.TurnIndex > 0 {
		turnFilter = query.TurnIndex
	}

	out := make([]types.AgentEvent, 0, len(events))
	for _, ev := range events {
		if query.RewindTo > 0 && ev.Sequence > query.RewindTo {
			continue
		}
		if len(typeSet) > 0 {
			if _, ok := typeSet[ev.Type]; !ok {
				continue
			}
		}
		if len(classSet) > 0 {
			if _, ok := classSet[eventClassFor(ev.Type)]; !ok {
				continue
			}
		}
		if turnFilter > 0 && ev.Turn != turnFilter {
			continue
		}
		if query.ReplayLabel != "" && ev.ReplayLabel != query.ReplayLabel {
			continue
		}
		if len(stopSet) > 0 {
			if _, ok := stopSet[ev.StopReason]; !ok {
				continue
			}
		}
		if query.ToolID != "" && ev.ToolUseID != query.ToolID {
			continue
		}
		if query.RequireTools && ev.ToolUseID == "" {
			continue
		}
		out = append(out, ev)
	}

	if query.Continuity == types.AgentReplayContinuityStrict {
		out = enforceStrictReplayContinuity(events, out)
	}

	truncated := false
	if query.Limit > 0 && len(out) > query.Limit {
		truncated = true
		out = out[len(out)-query.Limit:]
	}
	return out, truncated
}

func eventClassFor(eventType types.AgentEventType) types.AgentEventClass {
	switch eventType {
	case types.AgentEventTurnStart, types.AgentEventTurnEnd, types.AgentEventStateTransition, types.AgentEventTurnPhase, types.AgentEventStop:
		return types.AgentEventClassLifecycle
	case types.AgentEventReplayCheckpoint:
		return types.AgentEventClassReplay
	case types.AgentEventRetry:
		return types.AgentEventClassRetry
	case types.AgentEventCompactionStart, types.AgentEventCompactionEnd:
		return types.AgentEventClassCompaction
	case types.AgentEventToolQueued, types.AgentEventToolStart, types.AgentEventToolEnd, types.AgentEventToolRetry, types.AgentEventToolLifecycle:
		return types.AgentEventClassTool
	case types.AgentEventPermissionAsk, types.AgentEventPermissionResult:
		return types.AgentEventClassPermission
	case types.AgentEventTaskLifecycle, types.AgentEventTeamLifecycle:
		return types.AgentEventClassRuntime
	default:
		return types.AgentEventClassIO
	}
}
