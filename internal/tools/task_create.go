package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/alliecatowo/alliecode/internal/tasks"
	"github.com/alliecatowo/alliecode/internal/types"
)

// TaskCreateTool creates a local task record.
type TaskCreateTool struct{}

type taskCreateInput struct {
	Subject     string         `json:"subject"`
	Description string         `json:"description"`
	ActiveForm  string         `json:"activeForm,omitempty"`
	Owner       string         `json:"owner,omitempty"`
	Status      string         `json:"status,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

var taskCreateSeq atomic.Uint64

func (t *TaskCreateTool) Name() string { return "task_create" }

func (t *TaskCreateTool) Description() string {
	return "Creates a task in the local task list."
}

func (t *TaskCreateTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"subject":     {Type: "string", Description: "A brief title for the task."},
			"description": {Type: "string", Description: "What needs to be done."},
			"activeForm":  {Type: "string", Description: "Optional present-continuous verb phrase for in-progress display."},
			"metadata":    {Type: "object", Description: "Optional metadata to attach to the task."},
		},
		Required: []string{"subject", "description"},
	}
}

func (t *TaskCreateTool) Execute(_ context.Context, input types.ToolInput, _ types.ToolContext) (types.ToolResult, error) {
	var in taskCreateInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("invalid input: %v", err), IsError: true}, nil
	}

	in.Subject = strings.TrimSpace(in.Subject)
	in.Description = strings.TrimSpace(in.Description)
	in.ActiveForm = strings.TrimSpace(in.ActiveForm)
	in.Owner = strings.TrimSpace(in.Owner)
	if in.Subject == "" {
		return types.ToolResult{Content: "subject is required", IsError: true}, nil
	}
	if in.Description == "" {
		return types.ToolResult{Content: "description is required", IsError: true}, nil
	}
	status, ok := canonicalTaskStatus(strings.TrimSpace(strings.ToLower(in.Status)))
	if !ok {
		return types.ToolResult{Content: "status must be one of: pending, in_progress, completed, failed, canceled, deleted", IsError: true}, nil
	}
	if status == "" {
		status = taskStatusPending
	}

	now := time.Now().UTC()
	taskID := fmt.Sprintf("task-%d-%d", now.UnixNano(), taskCreateSeq.Add(1))
	managerStatus := managerStatusFromTaskStatus(status)
	rec := taskAdapterUpdate(taskID, string(managerStatus), "", "")
	rec = taskAdapterSetFields(taskID, func(r *taskAdapterRecord) {
		r.Subject = in.Subject
		r.Description = in.Description
		r.ActiveForm = in.ActiveForm
		r.Owner = in.Owner
		r.Metadata = cloneMetadata(in.Metadata)
	})
	task := tasks.Task{
		ID:          taskID,
		Subject:     in.Subject,
		Description: in.Description,
		ActiveForm:  in.ActiveForm,
		Owner:       in.Owner,
		Metadata:    cloneMetadata(in.Metadata),
		Status:      tasks.Status(status),
		CreatedAt:   rec.UpdatedAt,
		UpdatedAt:   rec.UpdatedAt,
	}

	summary := taskAdapterSummary(context.Background())
	runtime := map[string]any{"operation": "create", "requested_status": status}
	b, err := json.Marshal(taskEnvelope{Task: &task, Summary: summary, Lifecycle: summarizeTaskLifecycle(&task), Runtime: runtime})
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}

	return types.ToolResult{Content: string(b)}, nil
}

func (t *TaskCreateTool) IsReadOnly(input types.ToolInput) bool { return false }

func (t *TaskCreateTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *TaskCreateTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *TaskCreateTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyTask, "task_create", input, toolCtx, types.PermissionAllowed)
}
