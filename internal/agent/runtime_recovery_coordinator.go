package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/alliecatowo/alliecode/internal/types"
)

type recoveryCoordinator struct {
	agent *Agent
}

func newRecoveryCoordinator(agent *Agent) recoveryCoordinator {
	return recoveryCoordinator{agent: agent}
}

func (r recoveryCoordinator) RecoverPromptTooLong(ctx context.Context) (bool, error) {
	if ctx.Err() != nil {
		r.agent.emit(types.AgentEvent{
			Type: types.AgentEventRetry,
			Recovery: types.AgentRecoveryState{
				Kind:         "prompt_too_long",
				Branch:       "prompt_too_long",
				Recovered:    false,
				FromPhase:    types.AgentTurnPhaseProviderRequest,
				ToPhase:      types.AgentTurnPhaseRetry,
				Attempt:      1,
				MaxAttempts:  1,
				ContextError: ctx.Err().Error(),
			},
			Runtime: r.agent.RuntimeSnapshot(),
		})
		return false, ctx.Err()
	}
	compacted, compactErr := r.agent.attemptCompaction(ctx, true)
	if compactErr != nil {
		return false, compactErr
	}
	if !compacted {
		r.agent.emit(types.AgentEvent{
			Type: types.AgentEventRetry,
			Recovery: types.AgentRecoveryState{
				Kind:        "prompt_too_long",
				Branch:      "prompt_too_long",
				Recovered:   false,
				Compacted:   false,
				FromPhase:   types.AgentTurnPhaseProviderRequest,
				ToPhase:     types.AgentTurnPhaseRetry,
				Attempt:     1,
				MaxAttempts: 1,
			},
			Runtime: r.agent.RuntimeSnapshot(),
		})
		return false, nil
	}

	prompt := promptTooLongRecoveryMessage()
	if err := r.agent.appendMessage(types.NewTextMessage(types.RoleUser, prompt)); err != nil {
		return false, fmt.Errorf("persist prompt-too-long recovery message: %w", err)
	}
	if r.agent.config.Debug {
		log.Printf("[agent] recovered from provider too-long failure via forced compaction")
	}
	r.agent.emit(types.AgentEvent{
		Type: types.AgentEventRetry,
		Recovery: types.AgentRecoveryState{
			Kind:        "prompt_too_long",
			Branch:      "prompt_too_long",
			Recovered:   true,
			Compacted:   true,
			PromptText:  prompt,
			FromPhase:   types.AgentTurnPhaseProviderRequest,
			ToPhase:     types.AgentTurnPhaseRetry,
			Attempt:     1,
			MaxAttempts: 1,
		},
		Runtime: r.agent.RuntimeSnapshot(),
	})
	return true, nil
}
