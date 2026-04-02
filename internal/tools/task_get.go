package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/tasks"
	"github.com/alliecatowo/alliecode/internal/types"
)

// TaskGetTool returns a background task snapshot.
type TaskGetTool struct{}

type taskGetInput struct {
	TaskID    string `json:"task_id"`
	TaskIDAlt string `json:"taskId"`
}

func firstTaskID(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

func (t *TaskGetTool) Name() string { return "task_get" }

func (t *TaskGetTool) Description() string {
	return "Gets the current status for a task by id."
}

func (t *TaskGetTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"task_id": {Type: "string", Description: "Task identifier."},
		},
		Required: []string{"task_id"},
	}
}

func (t *TaskGetTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in taskGetInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("invalid input: %v", err), IsError: true}, nil
	}
	in.TaskID = firstTaskID(in.TaskID, in.TaskIDAlt)
	if in.TaskID == "" {
		return types.ToolResult{Content: "task_id is required", IsError: true}, nil
	}

	task, ok := taskAdapterGet(ctx, in.TaskID)
	if !ok {
		summary := tasksStatusSummaryFromList(nil)
		runtime := map[string]any{"requested_task_id": in.TaskID, "found": false}
		b, err := json.Marshal(taskEnvelope{Task: nil, Summary: summary, Runtime: runtime})
		if err != nil {
			return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
		}
		return types.ToolResult{Content: string(b)}, nil
	}
	task.Status = tasks.Status(taskStatusFromManagerStatus(task.Status))
	summary := tasksStatusSummaryFromList([]tasks.Task{task})

	runtime := map[string]any{"requested_task_id": in.TaskID, "found": true}
	b, err := json.Marshal(taskEnvelope{Task: &task, Summary: summary, Lifecycle: summarizeTaskLifecycle(&task), Runtime: runtime})
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b)}, nil
}

func (t *TaskGetTool) IsReadOnly(input types.ToolInput) bool { return true }

func (t *TaskGetTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *TaskGetTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *TaskGetTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyTask, "task_get", input, toolCtx, types.PermissionAllowed)
}
