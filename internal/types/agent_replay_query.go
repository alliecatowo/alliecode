package types

// AgentEventClass groups event types into broader replay categories.
type AgentEventClass string

const (
	AgentEventClassLifecycle  AgentEventClass = "lifecycle"
	AgentEventClassReplay     AgentEventClass = "replay"
	AgentEventClassRetry      AgentEventClass = "retry"
	AgentEventClassCompaction AgentEventClass = "compaction"
	AgentEventClassTool       AgentEventClass = "tool"
	AgentEventClassPermission AgentEventClass = "permission"
	AgentEventClassRuntime    AgentEventClass = "runtime"
	AgentEventClassIO         AgentEventClass = "io"
)

// AgentReplayQuery specifies replay selection criteria.
type AgentReplayQuery struct {
	FromSequence uint64                    `json:"from_sequence,omitempty"`
	ToSequence   uint64                    `json:"to_sequence,omitempty"`
	Types        []AgentEventType          `json:"types,omitempty"`
	Classes      []AgentEventClass         `json:"classes,omitempty"`
	ToolID       string                    `json:"tool_id,omitempty"`
	Turn         int                       `json:"turn,omitempty"`
	TurnIndex    int                       `json:"turn_index,omitempty"`
	ReplayLabel  string                    `json:"replay_label,omitempty"`
	StopReasons  []AgentStopReason         `json:"stop_reasons,omitempty"`
	RequireTools bool                      `json:"require_tools,omitempty"`
	Continuity   AgentReplayContinuityMode `json:"continuity,omitempty"`
	RewindTo     uint64                    `json:"rewind_to,omitempty"`
	Limit        int                       `json:"limit,omitempty"`
}

// AgentReplayContinuityMode controls how replay preserves event continuity.
type AgentReplayContinuityMode string

const (
	AgentReplayContinuityNone   AgentReplayContinuityMode = "none"
	AgentReplayContinuityStrict AgentReplayContinuityMode = "strict"
)

// AgentReplayWindow represents replay coverage bounds.
type AgentReplayWindow struct {
	Start uint64 `json:"start,omitempty"`
	End   uint64 `json:"end,omitempty"`
	Total int    `json:"total,omitempty"`
}

// AgentEventTypeCount captures count per event type in a replay view.
type AgentEventTypeCount struct {
	Type  AgentEventType `json:"type"`
	Count int            `json:"count"`
}

// AgentEventClassCount captures count per event class.
type AgentEventClassCount struct {
	Class AgentEventClass `json:"class"`
	Count int             `json:"count"`
}

// AgentReplayStats captures summary stats for a replay result.
type AgentReplayStats struct {
	Matched      int                    `json:"matched,omitempty"`
	Truncated    bool                   `json:"truncated,omitempty"`
	TypeCounts   []AgentEventTypeCount  `json:"type_counts,omitempty"`
	ClassCounts  []AgentEventClassCount `json:"class_counts,omitempty"`
	FirstSeq     uint64                 `json:"first_sequence,omitempty"`
	LastSeq      uint64                 `json:"last_sequence,omitempty"`
	TurnsSeen    int                    `json:"turns_seen,omitempty"`
	ToolsSeen    int                    `json:"tools_seen,omitempty"`
	HasStopEvent bool                   `json:"has_stop_event,omitempty"`
	Continuity   AgentReplayContinuity  `json:"continuity,omitempty"`
	Rewind       AgentReplayRewindState `json:"rewind,omitempty"`
}

// AgentReplayContinuity describes continuity metadata for a replay slice.
type AgentReplayContinuity struct {
	Mode              AgentReplayContinuityMode `json:"mode,omitempty"`
	Strict            bool                      `json:"strict,omitempty"`
	SequenceGaps      int                       `json:"sequence_gaps,omitempty"`
	MissingParentRefs int                       `json:"missing_parent_refs,omitempty"`
	LastReplayLabel   string                    `json:"last_replay_label,omitempty"`
}

// AgentReplayRewindState describes rewind checkpoint information in a replay view.
type AgentReplayRewindState struct {
	Applied      bool               `json:"applied,omitempty"`
	TargetSeq    uint64             `json:"target_sequence,omitempty"`
	TargetLabel  string             `json:"target_label,omitempty"`
	Checkpointed bool               `json:"checkpointed,omitempty"`
	Checkpoints  []AgentRewindPoint `json:"checkpoints,omitempty"`
}

// AgentRewindPoint identifies a deterministic rewind checkpoint.
type AgentRewindPoint struct {
	Sequence    uint64 `json:"sequence"`
	Turn        int    `json:"turn,omitempty"`
	ReplayLabel string `json:"replay_label,omitempty"`
}

// AgentContextDependencyKind identifies dependency node kinds.
type AgentContextDependencyKind string

const (
	AgentContextDependencyEvent    AgentContextDependencyKind = "event"
	AgentContextDependencyTurn     AgentContextDependencyKind = "turn"
	AgentContextDependencyToolUse  AgentContextDependencyKind = "tool_use"
	AgentContextDependencyToolFlow AgentContextDependencyKind = "tool_flow"
)

// AgentContextDependencyNode captures one node in replay dependency graph.
type AgentContextDependencyNode struct {
	ID          string                     `json:"id"`
	Kind        AgentContextDependencyKind `json:"kind"`
	Sequence    uint64                     `json:"sequence,omitempty"`
	Turn        int                        `json:"turn,omitempty"`
	ReplayLabel string                     `json:"replay_label,omitempty"`
	ToolUseID   string                     `json:"tool_use_id,omitempty"`
	EventType   AgentEventType             `json:"event_type,omitempty"`
	DependsOn   []string                   `json:"depends_on,omitempty"`
}

// AgentContextDependencyGraph is a replay dependency graph summary.
type AgentContextDependencyGraph struct {
	Nodes []AgentContextDependencyNode `json:"nodes,omitempty"`
}

// AgentReplayView is a typed replay query result.
type AgentReplayView struct {
	Query           AgentReplayQuery            `json:"query"`
	Window          AgentReplayWindow           `json:"window"`
	Cursor          AgentReplayCursor           `json:"cursor"`
	Stats           AgentReplayStats            `json:"stats"`
	DependencyGraph AgentContextDependencyGraph `json:"dependency_graph,omitempty"`
	Events          []AgentEvent                `json:"events"`
}
