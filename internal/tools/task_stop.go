package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/tasks"
	"github.com/alliecatowo/alliecode/internal/types"
)

// TaskStopTool marks a running task as failed.
type TaskStopTool struct{}

type taskStopInput struct {
	TaskID    string `json:"task_id"`
	TaskIDAlt string `json:"shell_id,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

func (t *TaskStopTool) Name() string { return "task_stop" }

func (t *TaskStopTool) Description() string {
	return "Stops a task and records a cancellation reason."
}

func (t *TaskStopTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"task_id": {Type: "string", Description: "Task identifier."},
			"reason":  {Type: "string", Description: "Optional stop reason."},
		},
		Required: []string{"task_id"},
	}
}

func (t *TaskStopTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in taskStopInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("invalid input: %v", err), IsError: true}, nil
	}
	in.TaskID = firstTaskID(in.TaskID, in.TaskIDAlt)
	if in.TaskID == "" {
		return types.ToolResult{Content: "task_id is required", IsError: true}, nil
	}
	task, ok := taskAdapterGet(ctx, in.TaskID)
	if !ok {
		return types.ToolResult{Content: fmt.Sprintf("task %q not found", in.TaskID), IsError: true}, nil
	}
	if task.Status != tasks.StatusRunning {
		return types.ToolResult{Content: fmt.Sprintf("task %q is not running (status: %s)", in.TaskID, task.Status), IsError: true}, nil
	}

	reason := strings.TrimSpace(in.Reason)
	if reason == "" {
		reason = "stopped"
	}

	if agentTaskManager != nil {
		if t, ok := agentTaskManager.Cancel(in.TaskID, reason); ok {
			runtime := map[string]any{"operation": "stop", "requested_task_id": in.TaskID, "reason": reason, "manager": "agent"}
			b, err := json.Marshal(taskEnvelope{Task: &t, Lifecycle: summarizeTaskLifecycle(&t), Runtime: runtime, Contract: defaultTaskContractMetadata()})
			if err != nil {
				return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
			}
			return types.ToolResult{Content: string(b)}, nil
		}
	}

	rec := taskAdapterUpdate(in.TaskID, string(tasks.StatusCanceled), "", reason)
	updatedTask := tasks.Task{
		ID:        in.TaskID,
		Status:    tasks.Status(taskStatusCanceled),
		Error:     rec.Error,
		CreatedAt: rec.UpdatedAt,
		UpdatedAt: rec.UpdatedAt,
	}
	runtime := map[string]any{"operation": "stop", "requested_task_id": in.TaskID, "reason": reason, "manager": "adapter"}
	b, err := json.Marshal(taskEnvelope{Task: &updatedTask, Lifecycle: summarizeTaskLifecycle(&updatedTask), Runtime: runtime, Contract: defaultTaskContractMetadata()})
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}

	return types.ToolResult{Content: string(b)}, nil
}

func (t *TaskStopTool) IsReadOnly(input types.ToolInput) bool { return false }

func (t *TaskStopTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *TaskStopTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *TaskStopTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyTask, "task_stop", input, toolCtx, types.PermissionAllowed)
}
