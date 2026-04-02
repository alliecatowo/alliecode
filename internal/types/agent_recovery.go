package types

// AgentRecoveryState summarizes recovery workflow outcomes.
type AgentRecoveryState struct {
	Kind         string         `json:"kind,omitempty"`
	Branch       string         `json:"branch,omitempty"`
	Recovered    bool           `json:"recovered,omitempty"`
	Compacted    bool           `json:"compacted,omitempty"`
	Details      string         `json:"details,omitempty"`
	PromptText   string         `json:"prompt_text,omitempty"`
	FromPhase    AgentTurnPhase `json:"from_phase,omitempty"`
	ToPhase      AgentTurnPhase `json:"to_phase,omitempty"`
	Attempt      int            `json:"attempt,omitempty"`
	MaxAttempts  int            `json:"max_attempts,omitempty"`
	ContextError string         `json:"context_error,omitempty"`
}
