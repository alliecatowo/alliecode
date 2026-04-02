package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/alliecatowo/alliecode/internal/types"
)

type compactionCoordinator struct {
	agent *Agent
}

func newCompactionCoordinator(agent *Agent) compactionCoordinator {
	return compactionCoordinator{agent: agent}
}

func (c compactionCoordinator) Attempt(ctx context.Context, force bool) (bool, error) {
	c.agent.emitTurnPhase(0, types.AgentTurnPhaseCompactionCheck, c.agent.RuntimeSnapshot().Phase, "compaction_attempt", false, "")
	c.agent.mu.Lock()
	beforeCount := len(c.agent.messages)
	c.agent.mu.Unlock()

	c.agent.emitTransition(0, types.AgentLifecycleCompacting, types.AgentTransitionCompactionCheck)
	c.agent.emit(types.AgentEvent{
		Type:                  types.AgentEventCompactionStart,
		CompactionForced:      force,
		CompactionBeforeCount: beforeCount,
		Compaction: types.AgentCompactionView{
			Forced:      force,
			BeforeCount: beforeCount,
		},
		Runtime: c.agent.RuntimeSnapshot(),
	})

	c.agent.mu.Lock()
	if !force && !ShouldCompact(c.agent.messages, c.agent.config.ContextWindow) {
		c.agent.mu.Unlock()
		c.agent.emitTransition(0, types.AgentLifecycleProviderRequest, types.AgentTransitionCompactionSkipped)
		c.agent.emit(types.AgentEvent{
			Type:                  types.AgentEventCompactionEnd,
			CompactionForced:      force,
			CompactionApplied:     false,
			CompactionBeforeCount: beforeCount,
			CompactionAfterCount:  beforeCount,
			Compaction: types.AgentCompactionView{
				Forced:      force,
				Applied:     false,
				BeforeCount: beforeCount,
				AfterCount:  beforeCount,
			},
			Runtime: c.agent.RuntimeSnapshot(),
		})
		return false, nil
	}

	compacted, boundary, err := CompactMessages(ctx, c.agent.messages, c.agent.config.Provider, c.agent.config.Model)
	if err != nil {
		c.agent.mu.Unlock()
		if c.agent.config.Debug {
			log.Printf("[agent] compaction failed: %v", err)
		}
		c.agent.emitTransition(0, types.AgentLifecycleProviderRequest, types.AgentTransitionCompactionSkipped)
		c.agent.emit(types.AgentEvent{
			Type:                  types.AgentEventCompactionEnd,
			CompactionForced:      force,
			CompactionApplied:     false,
			CompactionBeforeCount: beforeCount,
			CompactionAfterCount:  beforeCount,
			Details:               err.Error(),
			Compaction: types.AgentCompactionView{
				Forced:      force,
				Applied:     false,
				BeforeCount: beforeCount,
				AfterCount:  beforeCount,
			},
			Runtime: c.agent.RuntimeSnapshot(),
		})
		return false, nil
	}

	if boundary != nil && c.agent.session != nil {
		if persistErr := c.agent.session.AppendCompactionBoundary(*boundary); persistErr != nil {
			c.agent.mu.Unlock()
			return false, fmt.Errorf("persist compaction boundary: %w", persistErr)
		}
	}

	currentCount := len(c.agent.messages)
	if len(compacted) == currentCount {
		c.agent.mu.Unlock()
		c.agent.emitTransition(0, types.AgentLifecycleProviderRequest, types.AgentTransitionCompactionSkipped)
		c.agent.emit(types.AgentEvent{
			Type:                  types.AgentEventCompactionEnd,
			CompactionForced:      force,
			CompactionApplied:     false,
			CompactionBeforeCount: beforeCount,
			CompactionAfterCount:  currentCount,
			Compaction: types.AgentCompactionView{
				Forced:      force,
				Applied:     false,
				BeforeCount: beforeCount,
				AfterCount:  currentCount,
			},
			Runtime: c.agent.RuntimeSnapshot(),
		})
		return false, nil
	}

	c.agent.messages = compacted
	afterCount := len(compacted)
	c.agent.mu.Unlock()

	c.agent.emitTransition(0, types.AgentLifecycleProviderRequest, types.AgentTransitionCompactionApplied)
	c.agent.emit(types.AgentEvent{
		Type:                  types.AgentEventCompactionEnd,
		CompactionForced:      force,
		CompactionApplied:     true,
		CompactionBeforeCount: beforeCount,
		CompactionAfterCount:  afterCount,
		Compaction: types.AgentCompactionView{
			Forced:      force,
			Applied:     true,
			BeforeCount: beforeCount,
			AfterCount:  afterCount,
		},
		Runtime: c.agent.RuntimeSnapshot(),
	})

	return true, nil
}
