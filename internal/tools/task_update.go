package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/tasks"
	"github.com/alliecatowo/alliecode/internal/types"
)

// TaskUpdateTool updates local task adapter state.
type TaskUpdateTool struct{}

type taskUpdateInput struct {
	TaskID      string         `json:"task_id"`
	TaskIDAlt   string         `json:"taskId,omitempty"`
	Subject     string         `json:"subject,omitempty"`
	Description string         `json:"description,omitempty"`
	ActiveForm  string         `json:"activeForm,omitempty"`
	Owner       string         `json:"owner,omitempty"`
	Status      string         `json:"status,omitempty"`
	Result      string         `json:"result,omitempty"`
	Error       string         `json:"error,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

func (t *TaskUpdateTool) Name() string { return "task_update" }

func (t *TaskUpdateTool) Description() string {
	return "Updates task status/result/error metadata."
}

func (t *TaskUpdateTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"task_id":     {Type: "string", Description: "Task identifier."},
			"subject":     {Type: "string", Description: "Optional new subject."},
			"description": {Type: "string", Description: "Optional new description."},
			"activeForm":  {Type: "string", Description: "Optional present-continuous active form."},
			"owner":       {Type: "string", Description: "Optional owner identifier."},
			"status":      {Type: "string", Description: "Task status (pending, in_progress, completed, failed, canceled, deleted)."},
			"result":      {Type: "string", Description: "Optional task result text."},
			"error":       {Type: "string", Description: "Optional task error text."},
			"metadata":    {Type: "object", Description: "Optional metadata map."},
		},
		Required: []string{"task_id"},
	}
}

func (t *TaskUpdateTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in taskUpdateInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("invalid input: %v", err), IsError: true}, nil
	}
	in.TaskID = firstTaskID(in.TaskID, in.TaskIDAlt)
	if in.TaskID == "" {
		return types.ToolResult{Content: "task_id is required", IsError: true}, nil
	}

	if _, ok := taskAdapterGet(ctx, in.TaskID); !ok {
		return types.ToolResult{Content: fmt.Sprintf("task %q not found", in.TaskID), IsError: true}, nil
	}
	if !hasTaskUpdateFields(in) {
		return types.ToolResult{Content: "at least one updatable field is required", IsError: true}, nil
	}

	canonicalStatus, ok := canonicalTaskStatus(strings.TrimSpace(strings.ToLower(in.Status)))
	if !ok {
		return types.ToolResult{Content: "status must be one of: pending, in_progress, completed, failed, canceled, deleted", IsError: true}, nil
	}
	if canonicalStatus == taskStatusDeleted {
		taskAdapterMu.Lock()
		delete(taskAdapterRecords, in.TaskID)
		taskAdapterMu.Unlock()
		summary := taskAdapterSummary(context.Background())
		runtime := map[string]any{"operation": "delete", "requested_task_id": in.TaskID}
		b, err := json.Marshal(taskEnvelope{Task: nil, Summary: summary, Runtime: runtime, Contract: defaultTaskContractMetadata()})
		if err != nil {
			return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
		}
		return types.ToolResult{Content: string(b)}, nil
	}

	statusForUpdate := ""
	if canonicalStatus != "" {
		statusForUpdate = string(managerStatusFromTaskStatus(canonicalStatus))
	}
	rec := taskAdapterUpdate(in.TaskID, statusForUpdate, in.Result, in.Error)
	rec = taskAdapterSetFields(in.TaskID, func(r *taskAdapterRecord) {
		if strings.TrimSpace(in.Subject) != "" {
			r.Subject = strings.TrimSpace(in.Subject)
		}
		if strings.TrimSpace(in.Description) != "" {
			r.Description = strings.TrimSpace(in.Description)
		}
		if strings.TrimSpace(in.ActiveForm) != "" {
			r.ActiveForm = strings.TrimSpace(in.ActiveForm)
		}
		if strings.TrimSpace(in.Owner) != "" {
			r.Owner = strings.TrimSpace(in.Owner)
		}
		if in.Metadata != nil {
			merged := cloneMetadata(r.Metadata)
			if merged == nil {
				merged = map[string]any{}
			}
			for key, value := range in.Metadata {
				if value == nil {
					delete(merged, key)
					continue
				}
				merged[key] = value
			}
			if len(merged) == 0 {
				r.Metadata = nil
			} else {
				r.Metadata = merged
			}
		}
	})
	taskStatus := canonicalStatus
	if taskStatus == "" {
		taskStatus = taskStatusFromManagerStatus(tasks.Status(rec.Status))
	}
	task := tasks.Task{
		ID:          in.TaskID,
		Subject:     rec.Subject,
		Description: rec.Description,
		ActiveForm:  rec.ActiveForm,
		Owner:       rec.Owner,
		Metadata:    cloneMetadata(rec.Metadata),
		Status:      tasks.Status(taskStatus),
		Result:      rec.Result,
		Error:       rec.Error,
		CreatedAt:   rec.UpdatedAt,
		UpdatedAt:   rec.UpdatedAt,
	}
	summary := taskAdapterSummary(context.Background())
	runtime := map[string]any{"operation": "update", "requested_task_id": in.TaskID, "status": task.Status}
	b, err := json.Marshal(taskEnvelope{Task: &task, Summary: summary, Lifecycle: summarizeTaskLifecycle(&task), Runtime: runtime, Contract: defaultTaskContractMetadata()})
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}

	return types.ToolResult{Content: string(b)}, nil
}

func (t *TaskUpdateTool) IsReadOnly(input types.ToolInput) bool { return false }

func (t *TaskUpdateTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *TaskUpdateTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *TaskUpdateTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyTask, "task_update", input, toolCtx, types.PermissionAllowed)
}
