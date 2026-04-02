package agent

import (
	"context"
	"strconv"

	"github.com/alliecatowo/alliecode/internal/hooks"
	"github.com/alliecatowo/alliecode/internal/types"
)

const defaultMaxContinuationRecoveries = 2

type turnPlanner struct {
	agent *Agent
}

func newTurnPlanner(agent *Agent) turnPlanner {
	return turnPlanner{agent: agent}
}

func (p turnPlanner) BeginTurn(ctx context.Context, turnNumber int) error {
	p.agent.emitTransition(turnNumber, types.AgentLifecycleTurnStarting, types.AgentTransitionTurnStarted)
	p.agent.emit(types.AgentEvent{
		Type:    types.AgentEventTurnStart,
		Turn:    turnNumber,
		Runtime: p.agent.RuntimeSnapshot(),
		TurnPlan: types.AgentTurnPlan{
			Turn:            turnNumber,
			TurnIndex:       turnNumber,
			BudgetCheck:     true,
			CompactionCheck: true,
			Phase:           types.AgentTurnPhaseInit,
			ReplayLabel:     "turn_start",
		},
		ReplayLabel: "turn_start",
	})
	p.agent.emit(types.AgentEvent{Type: types.AgentEventReplayCheckpoint, Turn: turnNumber, ReplayLabel: "turn_start", Replay: p.agent.ReplayCursor()})
	p.agent.runtimeSetTurn(turnNumber)

	if p.agent.config.Hooks != nil {
		if err := p.agent.config.Hooks.Fire(ctx, hooks.EventTurnStart, map[string]string{
			"TURN": strconv.Itoa(turnNumber),
		}); err != nil {
			p.agent.emitStop(types.AgentStopHookError, "", err.Error(), turnNumber)
			return err
		}
	}

	select {
	case <-ctx.Done():
		p.agent.emitTurnPhase(turnNumber, types.AgentTurnPhaseInit, types.AgentTurnPhaseInit, "turn_start_canceled", true, "context_canceled")
		p.agent.emitStop(types.AgentStopCanceled, "", ctx.Err().Error(), turnNumber)
		return ctx.Err()
	default:
	}

	p.agent.emitTurnPhase(turnNumber, types.AgentTurnPhaseBudgetCheck, types.AgentTurnPhaseInit, "budget_check", false, "")
	if err := p.agent.checkBudgets(turnNumber); err != nil {
		return err
	}

	p.agent.emitTurnPhase(turnNumber, types.AgentTurnPhaseCompactionCheck, types.AgentTurnPhaseBudgetCheck, "compaction_check", false, "")
	if _, err := p.agent.attemptCompaction(ctx, false); err != nil {
		p.agent.emitStop(types.AgentStopPersistenceError, "", err.Error(), turnNumber)
		return err
	}

	return nil
}
