package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/alliecatowo/alliecode/internal/types"
)

// TodoWriteTool manages a task checklist.
type TodoWriteTool struct{}

type todoItem struct {
	Content    string `json:"content"`
	Status     string `json:"status"`
	ActiveForm string `json:"activeForm"`
}

type todoWriteInput struct {
	Todos []todoItem `json:"todos"`
}

// Global session-level todo storage.
var (
	globalTodos []todoItem
	todoMu      sync.Mutex
)

func (t *TodoWriteTool) Name() string { return "TodoWrite" }

func (t *TodoWriteTool) Description() string {
	return "Manages a task checklist. Create, update, and track todos for the current session."
}

func (t *TodoWriteTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"todos": {
				Type:        "array",
				Description: "Array of todo items with content, status, and optional activeForm fields.",
				Items: &types.PropertySchema{
					Type: "object",
				},
			},
		},
		Required: []string{"todos"},
	}
}

func (t *TodoWriteTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in todoWriteInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}

	todoMu.Lock()
	globalTodos = in.Todos
	todoMu.Unlock()

	return types.ToolResult{Content: formatTodos(in.Todos)}, nil
}

func formatTodos(todos []todoItem) string {
	if len(todos) == 0 {
		return "(no todos)"
	}

	var sb strings.Builder
	for i, item := range todos {
		if i > 0 {
			sb.WriteString("\n")
		}

		status := item.Status
		icon := "[ ]"
		switch strings.ToLower(status) {
		case "done", "completed", "complete":
			icon = "[x]"
		case "in_progress", "in-progress", "active":
			icon = "[~]"
		case "blocked":
			icon = "[!]"
		case "cancelled", "canceled":
			icon = "[-]"
		}

		sb.WriteString(fmt.Sprintf("%s %s", icon, item.Content))
		if item.ActiveForm != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", item.ActiveForm))
		}
	}

	return sb.String()
}

// GetTodos returns the current global todo list (for use by other components).
func GetTodos() []todoItem {
	todoMu.Lock()
	defer todoMu.Unlock()
	result := make([]todoItem, len(globalTodos))
	copy(result, globalTodos)
	return result
}

func (t *TodoWriteTool) IsReadOnly(input types.ToolInput) bool {
	return false
}

func (t *TodoWriteTool) IsDestructive(input types.ToolInput) bool {
	return false
}

func (t *TodoWriteTool) IsConcurrencySafe(input types.ToolInput) bool {
	return true // Protected by mutex.
}

func (t *TodoWriteTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}
