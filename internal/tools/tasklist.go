package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/tasks"
	"github.com/alliecatowo/alliecode/internal/types"
)

// TaskListTool returns the current session todo list.
type TaskListTool struct{}

func (t *TaskListTool) Name() string { return "task_list" }

func (t *TaskListTool) Description() string {
	return "Returns background tasks with status and metadata."
}

func (t *TaskListTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"status":           {Type: "string", Description: "Optional status filter (running, completed, failed, canceled).", Enum: []string{"running", "completed", "failed", "canceled"}},
			"owner":            {Type: "string", Description: "Optional owner filter (case-insensitive)."},
			"limit":            {Type: "integer", Description: "Optional max number of tasks to return."},
			"include_terminal": {Type: "boolean", Description: "Include completed/failed/canceled tasks. Defaults to true."},
		},
	}
}

func (t *TaskListTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var obj map[string]json.RawMessage
	if len(input) > 0 {
		if err := json.Unmarshal(input, &obj); err != nil {
			return types.ToolResult{Content: fmt.Sprintf("invalid input: %v", err), IsError: true}, nil
		}
	}

	statusFilter := ""
	ownerFilter := ""
	limit := 0
	includeTerminal := true
	if raw, ok := obj["status"]; ok {
		if err := json.Unmarshal(raw, &statusFilter); err != nil {
			return types.ToolResult{Content: "status must be a string", IsError: true}, nil
		}
		statusFilter = strings.ToLower(strings.TrimSpace(statusFilter))
		canonical, ok := canonicalTaskStatus(statusFilter)
		if !ok {
			return types.ToolResult{Content: "status must be one of: pending, in_progress, completed, failed, canceled, deleted", IsError: true}, nil
		}
		statusFilter = canonical
	}

	if raw, ok := obj["owner"]; ok {
		if err := json.Unmarshal(raw, &ownerFilter); err != nil {
			return types.ToolResult{Content: "owner must be a string", IsError: true}, nil
		}
	}
	ownerFilter = strings.TrimSpace(ownerFilter)

	if raw, ok := obj["limit"]; ok {
		if err := json.Unmarshal(raw, &limit); err != nil {
			return types.ToolResult{Content: "limit must be an integer", IsError: true}, nil
		}
		if limit < 0 {
			return types.ToolResult{Content: "limit must be >= 0", IsError: true}, nil
		}
	}

	if raw, ok := obj["include_terminal"]; ok {
		if err := json.Unmarshal(raw, &includeTerminal); err != nil {
			return types.ToolResult{Content: "include_terminal must be a boolean", IsError: true}, nil
		}
	}

	query := tasks.TaskQuery{Owner: ownerFilter, Limit: limit, IncludeTerminal: &includeTerminal}
	if statusFilter != "" {
		query.Statuses = []tasks.Status{managerStatusFromTaskStatus(statusFilter)}
	}

	taskItems, querySummary := tasks.ApplyTaskQuery(taskAdapterList(ctx), query)
	filtered := make([]tasks.Task, 0, len(taskItems))
	for _, task := range taskItems {
		if isTaskInternal(task) {
			continue
		}
		filtered = append(filtered, task)
	}
	taskItems = filtered
	querySummary.Returned = len(taskItems)
	summary := tasksStatusSummaryFromList(taskItems)
	lifecycle := make(map[string]taskLifecycleSummary, len(taskItems))
	for i := range taskItems {
		taskItems[i].Status = tasks.Status(taskStatusFromManagerStatus(taskItems[i].Status))
		if l := summarizeTaskLifecycle(&taskItems[i]); l != nil {
			lifecycle[taskItems[i].ID] = *l
		}
	}

	audit := &taskListAuditSummary{
		GeneratedAtUnixMs: time.Now().UTC().UnixMilli(),
		ReturnedCount:     len(taskItems),
		MatchedCount:      querySummary.Matched,
		HasQueryFilter:    statusFilter != "" || ownerFilter != "" || !includeTerminal,
	}
	b, err := json.Marshal(taskListEnvelope{Tasks: taskItems, Total: len(taskItems), Summary: &summary, Query: &query, QueryRun: &querySummary, Lifecycle: lifecycle, Audit: audit, Contract: defaultTaskContractMetadata()})
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}

	return types.ToolResult{Content: string(b)}, nil
}

func isTaskInternal(task tasks.Task) bool {
	if len(task.Metadata) == 0 {
		return false
	}
	raw, ok := task.Metadata["_internal"]
	if !ok {
		return false
	}
	flag, ok := raw.(bool)
	return ok && flag
}

func (t *TaskListTool) IsReadOnly(input types.ToolInput) bool { return true }

func (t *TaskListTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *TaskListTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *TaskListTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyTask, "task_list", input, toolCtx, types.PermissionAllowed)
}
