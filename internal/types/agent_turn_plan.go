package types

// AgentTurnPhase captures finer-grained orchestration phase boundaries.
type AgentTurnPhase string

const (
	AgentTurnPhaseInit            AgentTurnPhase = "init"
	AgentTurnPhaseBudgetCheck     AgentTurnPhase = "budget_check"
	AgentTurnPhaseCompactionCheck AgentTurnPhase = "compaction_check"
	AgentTurnPhaseProviderRequest AgentTurnPhase = "provider_request"
	AgentTurnPhaseProviderStream  AgentTurnPhase = "provider_stream"
	AgentTurnPhaseToolExecution   AgentTurnPhase = "tool_execution"
	AgentTurnPhaseRetry           AgentTurnPhase = "retry"
	AgentTurnPhaseTurnEnd         AgentTurnPhase = "turn_end"
)

// AgentTurnPlan captures planned turn orchestration actions.
type AgentTurnPlan struct {
	Turn             int            `json:"turn,omitempty"`
	TurnIndex        int            `json:"turn_index,omitempty"`
	BudgetCheck      bool           `json:"budget_check,omitempty"`
	CompactionCheck  bool           `json:"compaction_check,omitempty"`
	CompactionForced bool           `json:"compaction_forced,omitempty"`
	Phase            AgentTurnPhase `json:"phase,omitempty"`
	Canceled         bool           `json:"canceled,omitempty"`
	RecoveryBranch   string         `json:"recovery_branch,omitempty"`
	ReplayLabel      string         `json:"replay_label,omitempty"`
}

// AgentTurnPhaseState captures runtime phase transitions for a turn.
type AgentTurnPhaseState struct {
	Turn           int            `json:"turn,omitempty"`
	TurnIndex      int            `json:"turn_index,omitempty"`
	Phase          AgentTurnPhase `json:"phase,omitempty"`
	PreviousPhase  AgentTurnPhase `json:"previous_phase,omitempty"`
	Transition     string         `json:"transition,omitempty"`
	Canceled       bool           `json:"canceled,omitempty"`
	RecoveryBranch string         `json:"recovery_branch,omitempty"`
}
