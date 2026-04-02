package types

import "fmt"

var lifecycleTransitions = map[AgentLifecycleState]map[AgentLifecycleState]struct{}{
	AgentLifecycleIdle: {
		AgentLifecycleTurnStarting: {},
		AgentLifecycleStopped:      {},
	},
	AgentLifecycleTurnStarting: {
		AgentLifecycleProviderRequest: {},
		AgentLifecycleStopped:         {},
	},
	AgentLifecycleProviderRequest: {
		AgentLifecycleProviderStream: {},
		AgentLifecycleRetrying:       {},
		AgentLifecycleCompacting:     {},
		AgentLifecycleStopped:        {},
	},
	AgentLifecycleProviderStream: {
		AgentLifecycleToolExecution: {},
		AgentLifecycleRetrying:      {},
		AgentLifecycleCompacting:    {},
		AgentLifecycleTurnEnding:    {},
		AgentLifecycleStopped:       {},
	},
	AgentLifecycleToolExecution: {
		AgentLifecycleProviderRequest: {},
		AgentLifecycleTurnEnding:      {},
		AgentLifecycleRetrying:        {},
		AgentLifecycleStopped:         {},
	},
	AgentLifecycleCompacting: {
		AgentLifecycleProviderRequest: {},
		AgentLifecycleRetrying:        {},
		AgentLifecycleStopped:         {},
	},
	AgentLifecycleRetrying: {
		AgentLifecycleProviderRequest: {},
		AgentLifecycleProviderStream:  {},
		AgentLifecycleStopped:         {},
	},
	AgentLifecycleTurnEnding: {
		AgentLifecycleIdle:    {},
		AgentLifecycleStopped: {},
	},
	AgentLifecycleStopped: {},
}

// IsValidLifecycleTransition reports whether from->to is a supported state transition.
func IsValidLifecycleTransition(from, to AgentLifecycleState) bool {
	next, ok := lifecycleTransitions[from]
	if !ok {
		return false
	}
	_, ok = next[to]
	return ok
}

// ValidateLifecycleTransition validates a lifecycle transition payload.
func ValidateLifecycleTransition(tr AgentTransition) error {
	if tr.From == "" {
		return fmt.Errorf("transition.from is required")
	}
	if tr.To == "" {
		return fmt.Errorf("transition.to is required")
	}
	if tr.Reason == "" {
		return fmt.Errorf("transition.reason is required")
	}
	if !IsValidLifecycleTransition(tr.From, tr.To) {
		return fmt.Errorf("invalid lifecycle transition: %s -> %s", tr.From, tr.To)
	}
	return nil
}

var turnPhaseTransitions = map[AgentTurnPhase]map[AgentTurnPhase]struct{}{
	AgentTurnPhaseInit: {
		AgentTurnPhaseBudgetCheck:     {},
		AgentTurnPhaseCompactionCheck: {},
		AgentTurnPhaseProviderRequest: {},
	},
	AgentTurnPhaseBudgetCheck: {
		AgentTurnPhaseCompactionCheck: {},
		AgentTurnPhaseProviderRequest: {},
		AgentTurnPhaseTurnEnd:         {},
	},
	AgentTurnPhaseCompactionCheck: {
		AgentTurnPhaseProviderRequest: {},
		AgentTurnPhaseRetry:           {},
	},
	AgentTurnPhaseProviderRequest: {
		AgentTurnPhaseProviderStream: {},
		AgentTurnPhaseRetry:          {},
	},
	AgentTurnPhaseProviderStream: {
		AgentTurnPhaseToolExecution: {},
		AgentTurnPhaseRetry:         {},
		AgentTurnPhaseTurnEnd:       {},
	},
	AgentTurnPhaseToolExecution: {
		AgentTurnPhaseProviderRequest: {},
		AgentTurnPhaseRetry:           {},
		AgentTurnPhaseTurnEnd:         {},
	},
	AgentTurnPhaseRetry: {
		AgentTurnPhaseProviderRequest: {},
		AgentTurnPhaseProviderStream:  {},
		AgentTurnPhaseTurnEnd:         {},
	},
	AgentTurnPhaseTurnEnd: {},
}

// IsValidTurnPhaseTransition reports whether previous->next is valid.
func IsValidTurnPhaseTransition(previous, next AgentTurnPhase) bool {
	allowed, ok := turnPhaseTransitions[previous]
	if !ok {
		return false
	}
	_, ok = allowed[next]
	return ok
}

// ValidateTurnPhaseState validates an emitted turn-phase transition payload.
func ValidateTurnPhaseState(state AgentTurnPhaseState) error {
	if state.Phase == "" {
		return fmt.Errorf("turn_phase.phase is required")
	}
	if state.PreviousPhase == "" {
		return nil
	}
	if !IsValidTurnPhaseTransition(state.PreviousPhase, state.Phase) {
		return fmt.Errorf("invalid turn phase transition: %s -> %s", state.PreviousPhase, state.Phase)
	}
	return nil
}
