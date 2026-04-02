package types

// AgentRetryPlan describes retry orchestration state.
type AgentRetryPlan struct {
	Kind      string         `json:"kind,omitempty"`
	Attempt   int            `json:"attempt,omitempty"`
	Max       int            `json:"max,omitempty"`
	FromPhase AgentTurnPhase `json:"from_phase,omitempty"`
	ToPhase   AgentTurnPhase `json:"to_phase,omitempty"`
	Canceled  bool           `json:"canceled,omitempty"`
}
