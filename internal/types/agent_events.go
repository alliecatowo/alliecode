package types

import "encoding/json"

// AgentEventType identifies structured loop progress events.
type AgentEventType string

const (
	AgentEventAssistantChunk   AgentEventType = "assistant_chunk"
	AgentEventInputProcessed   AgentEventType = "input_processed"
	AgentEventSessionResumed   AgentEventType = "session_resumed"
	AgentEventStateTransition  AgentEventType = "state_transition"
	AgentEventTurnPhase        AgentEventType = "turn_phase"
	AgentEventReplayCheckpoint AgentEventType = "replay_checkpoint"
	AgentEventTurnStart        AgentEventType = "turn_start"
	AgentEventTurnEnd          AgentEventType = "turn_end"
	AgentEventCompactionStart  AgentEventType = "compaction_start"
	AgentEventCompactionEnd    AgentEventType = "compaction_end"
	AgentEventRetry            AgentEventType = "retry"
	AgentEventToolQueued       AgentEventType = "tool_queued"
	AgentEventToolStart        AgentEventType = "tool_start"
	AgentEventToolEnd          AgentEventType = "tool_end"
	AgentEventToolRetry        AgentEventType = "tool_retry"
	AgentEventToolLifecycle    AgentEventType = "tool_lifecycle"
	AgentEventTaskLifecycle    AgentEventType = "task_lifecycle"
	AgentEventTeamLifecycle    AgentEventType = "team_lifecycle"
	AgentEventPermissionAsk    AgentEventType = "permission_request"
	AgentEventPermissionResult AgentEventType = "permission_result"
	AgentEventStop             AgentEventType = "stop"
)

// AgentLifecycleState captures coarse loop phases.
type AgentLifecycleState string

const (
	AgentLifecycleIdle            AgentLifecycleState = "idle"
	AgentLifecycleTurnStarting    AgentLifecycleState = "turn_starting"
	AgentLifecycleProviderRequest AgentLifecycleState = "provider_request"
	AgentLifecycleProviderStream  AgentLifecycleState = "provider_stream"
	AgentLifecycleToolExecution   AgentLifecycleState = "tool_execution"
	AgentLifecycleCompacting      AgentLifecycleState = "compacting"
	AgentLifecycleRetrying        AgentLifecycleState = "retrying"
	AgentLifecycleTurnEnding      AgentLifecycleState = "turn_ending"
	AgentLifecycleStopped         AgentLifecycleState = "stopped"
)

// AgentTransitionReason records why the lifecycle moved to a new state.
type AgentTransitionReason string

const (
	AgentTransitionInputAccepted      AgentTransitionReason = "input_accepted"
	AgentTransitionTurnStarted        AgentTransitionReason = "turn_started"
	AgentTransitionProviderCall       AgentTransitionReason = "provider_call"
	AgentTransitionProviderChunk      AgentTransitionReason = "provider_chunk"
	AgentTransitionToolUseDetected    AgentTransitionReason = "tool_use_detected"
	AgentTransitionToolExecutionStart AgentTransitionReason = "tool_execution_start"
	AgentTransitionToolExecutionEnd   AgentTransitionReason = "tool_execution_end"
	AgentTransitionCompactionCheck    AgentTransitionReason = "compaction_check"
	AgentTransitionCompactionApplied  AgentTransitionReason = "compaction_applied"
	AgentTransitionCompactionSkipped  AgentTransitionReason = "compaction_skipped"
	AgentTransitionRetryPromptTooLong AgentTransitionReason = "retry_prompt_too_long"
	AgentTransitionRetryMaxTokens     AgentTransitionReason = "retry_max_tokens"
	AgentTransitionRetryProviderError AgentTransitionReason = "retry_provider_error"
	AgentTransitionRetryCanceled      AgentTransitionReason = "retry_canceled"
	AgentTransitionTurnCompleted      AgentTransitionReason = "turn_completed"
	AgentTransitionStopped            AgentTransitionReason = "stopped"
	AgentTransitionResumed            AgentTransitionReason = "resumed"
)

// AgentTransition is emitted for lifecycle transitions.
type AgentTransition struct {
	From   AgentLifecycleState   `json:"from,omitempty"`
	To     AgentLifecycleState   `json:"to,omitempty"`
	Reason AgentTransitionReason `json:"reason,omitempty"`
}

// AgentReplayCursor allows deterministic event replay for observers.
type AgentReplayCursor struct {
	Start uint64 `json:"start,omitempty"`
	End   uint64 `json:"end,omitempty"`
	Size  int    `json:"size,omitempty"`
	Label string `json:"label,omitempty"`
}

// AgentReference captures one resolved @reference from user input.
type AgentReference struct {
	Raw           string `json:"raw"`
	Path          string `json:"path"`
	CanonicalPath string `json:"canonical_path"`
	Line          int    `json:"line,omitempty"`
	Start         int    `json:"start"`
	End           int    `json:"end"`
	Exists        bool   `json:"exists"`
	ResourceType  string `json:"resource_type"`
}

// AgentStopReason indicates why the agent loop ended.
type AgentStopReason string

const (
	AgentStopEndTurn            AgentStopReason = "end_turn"
	AgentStopMaxTokens          AgentStopReason = "max_tokens"
	AgentStopContextLimit       AgentStopReason = "context_window_exceeded"
	AgentStopStopSeq            AgentStopReason = "stop_sequence"
	AgentStopContentFilter      AgentStopReason = "content_filter"
	AgentStopRefusal            AgentStopReason = "refusal"
	AgentStopMaxTurns           AgentStopReason = "max_turns"
	AgentStopBudgetUSD          AgentStopReason = "budget_usd"
	AgentStopBudgetToken        AgentStopReason = "budget_tokens"
	AgentStopCanceled           AgentStopReason = "context_canceled"
	AgentStopInputError         AgentStopReason = "input_error"
	AgentStopProviderError      AgentStopReason = "provider_error"
	AgentStopToolExecutionError AgentStopReason = "tool_execution_error"
	AgentStopToolUseMalformed   AgentStopReason = "tool_use_malformed"
	AgentStopHookError          AgentStopReason = "hook_error"
	AgentStopPersistenceError   AgentStopReason = "persistence_error"
	AgentStopMaxTokensExhausted AgentStopReason = "max_tokens_recovery_exhausted"
	AgentStopUnknown            AgentStopReason = "unknown"
)

// AgentRuntimeSnapshot carries loop runtime counters useful for statusline/status views.
type AgentRuntimeSnapshot struct {
	Turns          int             `json:"turns,omitempty"`
	TurnIndex      int             `json:"turn_index,omitempty"`
	ToolInflight   int             `json:"tool_inflight,omitempty"`
	Phase          AgentTurnPhase  `json:"phase,omitempty"`
	TasksTotal     int             `json:"tasks_total,omitempty"`
	TasksRunning   int             `json:"tasks_running,omitempty"`
	TasksCompleted int             `json:"tasks_completed,omitempty"`
	TeamsTotal     int             `json:"teams_total,omitempty"`
	TeamsActive    int             `json:"teams_active,omitempty"`
	LastStopReason AgentStopReason `json:"last_stop_reason,omitempty"`
}

// AgentPermissionDecision is a normalized permission result used in events.
type AgentPermissionDecision string

const (
	AgentPermissionAllow AgentPermissionDecision = "allow"
	AgentPermissionDeny  AgentPermissionDecision = "deny"
	AgentPermissionAsk   AgentPermissionDecision = "ask"
)

// AgentEvent is a structured event emitted by the agent loop.
type AgentEvent struct {
	ID          string `json:"id,omitempty"`
	Sequence    uint64 `json:"sequence,omitempty"`
	TimestampMS int64  `json:"timestamp_ms,omitempty"`

	Type       AgentEventType      `json:"type"`
	Turn       int                 `json:"turn,omitempty"`
	State      AgentLifecycleState `json:"state,omitempty"`
	Transition AgentTransition     `json:"transition,omitempty"`
	Replay     AgentReplayCursor   `json:"replay,omitempty"`
	TurnPlan   AgentTurnPlan       `json:"turn_plan,omitempty"`
	TurnPhase  AgentTurnPhaseState `json:"turn_phase,omitempty"`
	RetryPlan  AgentRetryPlan      `json:"retry_plan,omitempty"`
	Recovery   AgentRecoveryState  `json:"recovery,omitempty"`
	Compaction AgentCompactionView `json:"compaction,omitempty"`

	AssistantChunk string `json:"assistant_chunk,omitempty"`
	Input          string `json:"input,omitempty"`

	ResolvedReferences []AgentReference `json:"resolved_references,omitempty"`

	SessionID              string `json:"session_id,omitempty"`
	SessionPath            string `json:"session_path,omitempty"`
	SessionMessageCount    int    `json:"session_message_count,omitempty"`
	SessionBoundaryCount   int    `json:"session_boundary_count,omitempty"`
	SessionResumedFromPath bool   `json:"session_resumed_from_path,omitempty"`

	ToolUseID string          `json:"tool_use_id,omitempty"`
	ToolName  string          `json:"tool_name,omitempty"`
	ToolInput json.RawMessage `json:"tool_input,omitempty"`
	ToolError string          `json:"tool_error,omitempty"`

	ToolTelemetryID string `json:"tool_telemetry_id,omitempty"`
	ToolStartedAtMS int64  `json:"tool_started_at_ms,omitempty"`
	ToolEndedAtMS   int64  `json:"tool_ended_at_ms,omitempty"`
	ToolDurationMS  int64  `json:"tool_duration_ms,omitempty"`
	ToolAttempt     int    `json:"tool_attempt,omitempty"`
	ToolMaxAttempts int    `json:"tool_max_attempts,omitempty"`
	ToolQueueIndex  int    `json:"tool_queue_index,omitempty"`
	ToolQueueTotal  int    `json:"tool_queue_total,omitempty"`
	ToolLifecycle   string `json:"tool_lifecycle,omitempty"`

	CompactionForced      bool `json:"compaction_forced,omitempty"`
	CompactionApplied     bool `json:"compaction_applied,omitempty"`
	CompactionBeforeCount int  `json:"compaction_before_count,omitempty"`
	CompactionAfterCount  int  `json:"compaction_after_count,omitempty"`

	RetryKind    string `json:"retry_kind,omitempty"`
	RetryAttempt int    `json:"retry_attempt,omitempty"`
	RetryMax     int    `json:"retry_max,omitempty"`

	ReplayLabel string `json:"replay_label,omitempty"`

	TaskID     string `json:"task_id,omitempty"`
	TaskStatus string `json:"task_status,omitempty"`
	TeamName   string `json:"team_name,omitempty"`
	TeamStatus string `json:"team_status,omitempty"`

	Runtime AgentRuntimeSnapshot `json:"runtime,omitempty"`

	PermissionDecision AgentPermissionDecision `json:"permission_decision,omitempty"`

	StopReason         AgentStopReason `json:"stop_reason,omitempty"`
	ProviderStopReason StopReason      `json:"provider_stop_reason,omitempty"`
	Details            string          `json:"details,omitempty"`
}
