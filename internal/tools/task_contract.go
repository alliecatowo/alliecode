package tools

import (
	"time"

	"github.com/alliecatowo/alliecode/internal/tasks"
)

type taskEnvelope struct {
	Task      *tasks.Task           `json:"task"`
	Summary   any                   `json:"summary,omitempty"`
	Lifecycle *taskLifecycleSummary `json:"lifecycle,omitempty"`
	Runtime   any                   `json:"runtime,omitempty"`
	Contract  *taskContractMetadata `json:"contract,omitempty"`
}

type taskContractMetadata struct {
	Family          string `json:"family"`
	SchemaVersion   string `json:"schema_version"`
	StateBacked     bool   `json:"state_backed"`
	HasRuntime      bool   `json:"has_runtime"`
	HasLifecycle    bool   `json:"has_lifecycle"`
	PermissionAware bool   `json:"permission_aware"`
}

type taskListEnvelope struct {
	Tasks     []tasks.Task                    `json:"tasks"`
	Total     int                             `json:"total"`
	Summary   *tasks.StatusSummary            `json:"summary,omitempty"`
	Query     *tasks.TaskQuery                `json:"query,omitempty"`
	QueryRun  *tasks.TaskQuerySummary         `json:"query_summary,omitempty"`
	Lifecycle map[string]taskLifecycleSummary `json:"lifecycle,omitempty"`
	Audit     *taskListAuditSummary           `json:"audit,omitempty"`
	Contract  *taskContractMetadata           `json:"contract,omitempty"`
}

type taskOutputEnvelope struct {
	RetrievalStatus string                `json:"retrieval_status"`
	OutputStatus    string                `json:"output_status,omitempty"`
	Task            *tasks.Task           `json:"task"`
	Summary         any                   `json:"summary,omitempty"`
	Lifecycle       *taskLifecycleSummary `json:"lifecycle,omitempty"`
	Runtime         any                   `json:"runtime,omitempty"`
	Contract        *taskContractMetadata `json:"contract,omitempty"`
}

type taskRuntimeSummary struct {
	Source      string `json:"source"`
	Blocked     bool   `json:"blocked"`
	PollMs      int    `json:"poll_ms,omitempty"`
	TimeoutMs   int    `json:"timeout_ms,omitempty"`
	DurationMs  int64  `json:"duration_ms"`
	Found       bool   `json:"found"`
	Terminal    bool   `json:"terminal"`
	RequestedID string `json:"requested_task_id"`
	WaitedMs    int64  `json:"waited_ms,omitempty"`
	WaitLoops   int    `json:"wait_loops,omitempty"`
	PollUsedMs  int    `json:"poll_used_ms,omitempty"`
}

type taskLifecycleSummary struct {
	CreatedAtUnixMs int64 `json:"created_at_unix_ms,omitempty"`
	UpdatedAtUnixMs int64 `json:"updated_at_unix_ms,omitempty"`
	AgeMs           int64 `json:"age_ms,omitempty"`
	Terminal        bool  `json:"terminal"`
	HasResult       bool  `json:"has_result"`
	HasError        bool  `json:"has_error"`
	HistoryEvents   int   `json:"history_events,omitempty"`
}

type taskListAuditSummary struct {
	GeneratedAtUnixMs int64 `json:"generated_at_unix_ms"`
	ReturnedCount     int   `json:"returned_count"`
	MatchedCount      int   `json:"matched_count"`
	HasQueryFilter    bool  `json:"has_query_filter"`
}

func summarizeTaskLifecycle(task *tasks.Task) *taskLifecycleSummary {
	if task == nil {
		return nil
	}
	now := time.Now().UTC()
	out := &taskLifecycleSummary{
		Terminal:      task.Status == tasks.StatusCompleted || task.Status == tasks.StatusFailed || task.Status == tasks.StatusCanceled,
		HasResult:     task.Result != "",
		HasError:      task.Error != "",
		HistoryEvents: len(task.History),
	}
	if !task.CreatedAt.IsZero() {
		out.CreatedAtUnixMs = task.CreatedAt.UnixMilli()
		out.AgeMs = now.Sub(task.CreatedAt).Milliseconds()
	}
	if !task.UpdatedAt.IsZero() {
		out.UpdatedAtUnixMs = task.UpdatedAt.UnixMilli()
	}
	if out.AgeMs < 0 {
		out.AgeMs = 0
	}
	return out
}

const (
	taskStatusPending    = "pending"
	taskStatusInProgress = "in_progress"
	taskStatusCompleted  = "completed"
	taskStatusCanceled   = "canceled"
	taskStatusFailed     = "failed"
	taskStatusDeleted    = "deleted"
)

func canonicalTaskStatus(raw string) (string, bool) {
	switch raw {
	case "", taskStatusPending, taskStatusInProgress, taskStatusCompleted, taskStatusCanceled, taskStatusFailed, taskStatusDeleted:
		return raw, true
	case "running":
		return taskStatusInProgress, true
	default:
		return "", false
	}
}

func managerStatusFromTaskStatus(taskStatus string) tasks.Status {
	if taskStatus == "" || taskStatus == taskStatusPending || taskStatus == taskStatusInProgress {
		return tasks.StatusRunning
	}
	if taskStatus == taskStatusCompleted {
		return tasks.StatusCompleted
	}
	if taskStatus == taskStatusCanceled {
		return tasks.StatusCanceled
	}
	return tasks.StatusFailed
}

func taskStatusFromManagerStatus(status tasks.Status) string {
	if status == tasks.StatusCompleted {
		return taskStatusCompleted
	}
	if status == tasks.StatusCanceled {
		return taskStatusCanceled
	}
	if status == tasks.StatusFailed {
		return taskStatusFailed
	}
	return taskStatusInProgress
}

func defaultTaskContractMetadata() *taskContractMetadata {
	return &taskContractMetadata{
		Family:          "task",
		SchemaVersion:   "v2",
		StateBacked:     true,
		HasRuntime:      true,
		HasLifecycle:    true,
		PermissionAware: true,
	}
}
