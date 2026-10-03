package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alliecatowo/alliecode/internal/tasks"
	"github.com/alliecatowo/alliecode/internal/types"
)

// TaskOutputTool returns task output metadata, optionally blocking.
type TaskOutputTool struct{}

type taskOutputInput struct {
	TaskID       string `json:"task_id"`
	TaskIDAlt    string `json:"taskId,omitempty"`
	Block        *bool  `json:"block,omitempty"`
	TimeoutMs    int    `json:"timeout_ms"`
	Timeout      int    `json:"timeout,omitempty"`
	PollMs       int    `json:"poll_ms"`
	PollInterval int    `json:"poll_interval_ms,omitempty"`
}

func (t *TaskOutputTool) Name() string { return "task_output" }

func (t *TaskOutputTool) Description() string {
	return "Returns task status/result; can wait for completion."
}

func (t *TaskOutputTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"task_id":    {Type: "string", Description: "Task identifier."},
			"block":      {Type: "boolean", Description: "Wait until task reaches terminal state."},
			"timeout_ms": {Type: "integer", Description: "Maximum wait in milliseconds when block=true."},
			"poll_ms":    {Type: "integer", Description: "Polling interval in milliseconds."},
		},
		Required: []string{"task_id"},
	}
}

func (t *TaskOutputTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	started := time.Now()
	var in taskOutputInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("invalid input: %v", err), IsError: true}, nil
	}
	in.TaskID = firstTaskID(in.TaskID, in.TaskIDAlt)
	if in.TaskID == "" {
		return types.ToolResult{Content: "task_id is required", IsError: true}, nil
	}
	timeoutMs := in.TimeoutMs
	if timeoutMs == 0 {
		timeoutMs = in.Timeout
	}
	if timeoutMs < 0 {
		return types.ToolResult{Content: "timeout_ms must be >= 0", IsError: true}, nil
	}
	pollMs := in.PollMs
	if pollMs == 0 {
		pollMs = in.PollInterval
	}
	if pollMs < 0 {
		return types.ToolResult{Content: "poll_ms must be >= 0", IsError: true}, nil
	}
	block := false
	if in.Block != nil {
		block = *in.Block
	}

	poll := 200 * time.Millisecond
	if pollMs > 0 {
		poll = time.Duration(pollMs) * time.Millisecond
	}

	var task tasks.Task
	var err error
	waitLoops := 0
	if block {
		waitCtx := ctx
		cancel := func() {}
		if timeoutMs > 0 {
			waitCtx, cancel = context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
		}
		defer cancel()
		task, err = taskAdapterWait(waitCtx, in.TaskID, poll)
		if err != nil {
			if timeoutMs > 0 && waitCtx.Err() == context.DeadlineExceeded {
				runtime := taskRuntimeSummary{Source: "adapter", Blocked: true, PollMs: pollMs, PollUsedMs: int(poll.Milliseconds()), TimeoutMs: timeoutMs, DurationMs: time.Since(started).Milliseconds(), RequestedID: in.TaskID, WaitedMs: time.Since(started).Milliseconds(), WaitLoops: waitLoops}
				b, _ := json.Marshal(taskOutputEnvelope{RetrievalStatus: "timeout", OutputStatus: "running", Task: nil, Runtime: runtime, Contract: defaultTaskContractMetadata()})
				return types.ToolResult{Content: string(b)}, nil
			}
			return types.ToolResult{Content: fmt.Sprintf("task wait error: %v", err), IsError: true}, nil
		}
	} else {
		var ok bool
		task, ok = taskAdapterGet(ctx, in.TaskID)
		if !ok {
			runtime := taskRuntimeSummary{Source: "adapter", Blocked: false, PollMs: pollMs, PollUsedMs: int(poll.Milliseconds()), TimeoutMs: timeoutMs, DurationMs: time.Since(started).Milliseconds(), RequestedID: in.TaskID}
			b, _ := json.Marshal(taskOutputEnvelope{RetrievalStatus: "not_found", OutputStatus: "empty", Task: nil, Runtime: runtime, Contract: defaultTaskContractMetadata()})
			return types.ToolResult{Content: string(b)}, nil
		}
	}

	retrieval := "success"
	outputStatus := "completed"
	taskState := taskStatusFromManagerStatus(task.Status)
	task.Status = tasks.Status(taskState)
	switch taskState {
	case taskStatusPending, taskStatusInProgress:
		retrieval = "not_ready"
		outputStatus = "running"
	case taskStatusFailed:
		outputStatus = "completed"
	case taskStatusCompleted, taskStatusCanceled:
		outputStatus = "completed"
	default:
		outputStatus = "completed"
	}

	summary := tasksStatusSummaryFromList([]tasks.Task{task})
	durationMs := time.Since(started).Milliseconds()
	if poll > 0 {
		waitLoops = int(durationMs / poll.Milliseconds())
	}
	runtime := taskRuntimeSummary{Source: "adapter", Blocked: block, PollMs: pollMs, PollUsedMs: int(poll.Milliseconds()), TimeoutMs: timeoutMs, DurationMs: durationMs, RequestedID: in.TaskID, Found: true, Terminal: taskState == taskStatusCompleted || taskState == taskStatusFailed || taskState == taskStatusCanceled, WaitedMs: durationMs, WaitLoops: waitLoops}
	b, err := json.Marshal(taskOutputEnvelope{RetrievalStatus: retrieval, OutputStatus: outputStatus, Task: &task, Summary: summary, Lifecycle: summarizeTaskLifecycle(&task), Runtime: runtime, Contract: defaultTaskContractMetadata()})
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b), IsError: taskState == taskStatusFailed}, nil
}

func (t *TaskOutputTool) IsReadOnly(input types.ToolInput) bool { return true }

func (t *TaskOutputTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *TaskOutputTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *TaskOutputTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyTask, "task_output", input, toolCtx, types.PermissionAllowed)
}
