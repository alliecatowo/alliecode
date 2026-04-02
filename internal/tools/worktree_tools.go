package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

type EnterWorktreeTool struct{}
type ExitWorktreeTool struct{}

var worktreeNamePattern = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,64}$`)

type enterWorktreeInput struct {
	Name string `json:"name,omitempty"`
}

type enterWorktreeOutput struct {
	Success      bool   `json:"success"`
	WorktreeName string `json:"worktree_name,omitempty"`
	WorktreePath string `json:"worktree_path,omitempty"`
	Message      string `json:"message"`
	Error        string `json:"error,omitempty"`
}

type exitWorktreeInput struct {
	Action         string `json:"action,omitempty"`
	DiscardChanges bool   `json:"discard_changes,omitempty"`
}

type exitWorktreeOutput struct {
	Success      bool   `json:"success"`
	Action       string `json:"action"`
	WorktreeName string `json:"worktree_name,omitempty"`
	WorktreePath string `json:"worktree_path,omitempty"`
	Removed      bool   `json:"removed"`
	Message      string `json:"message"`
	Error        string `json:"error,omitempty"`
}

func (t *EnterWorktreeTool) Name() string { return "worktree_enter" }

func (t *EnterWorktreeTool) Description() string {
	return "Creates a local isolated worktree directory and marks it as active."
}

func (t *EnterWorktreeTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"name": {Type: "string", Description: "Optional worktree name (letters, digits, dot, underscore, dash)."},
		},
	}
}

func (t *EnterWorktreeTool) Execute(_ context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in enterWorktreeInput
	if err := json.Unmarshal(input, &in); err != nil {
		return enterWorktreeResult(enterWorktreeOutput{Success: false, Message: "failed", Error: fmt.Sprintf("invalid input: %v", err)}, true)
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = "default"
	}
	if !worktreeNamePattern.MatchString(name) {
		return enterWorktreeResult(enterWorktreeOutput{Success: false, Message: "failed", Error: "invalid worktree name"}, true)
	}

	orchestrationState.mu.Lock()
	defer orchestrationState.mu.Unlock()
	if orchestrationState.worktree != nil {
		return enterWorktreeResult(enterWorktreeOutput{Success: false, Message: "failed", Error: "already in a worktree session"}, true)
	}

	base := filepath.Join(toolCtx.WorkingDir, ".alliecode-worktrees")
	if err := os.MkdirAll(base, 0o755); err != nil {
		return enterWorktreeResult(enterWorktreeOutput{Success: false, Message: "failed", Error: err.Error()}, true)
	}
	path := filepath.Join(base, name)
	if _, err := os.Stat(path); err == nil {
		for i := 2; i < 1000; i++ {
			candidate := fmt.Sprintf("%s-%d", name, i)
			candidatePath := filepath.Join(base, candidate)
			if _, e := os.Stat(candidatePath); os.IsNotExist(e) {
				name = candidate
				path = candidatePath
				break
			}
		}
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return enterWorktreeResult(enterWorktreeOutput{Success: false, Message: "failed", Error: err.Error()}, true)
	}

	orchestrationState.worktree = &worktreeSession{
		Name:        name,
		OriginalDir: toolCtx.WorkingDir,
		Path:        path,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	return enterWorktreeResult(enterWorktreeOutput{Success: true, WorktreeName: name, WorktreePath: path, Message: "worktree session created"}, false)
}

func enterWorktreeResult(out enterWorktreeOutput, isError bool) (types.ToolResult, error) {
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b), IsError: isError}, nil
}

func (t *EnterWorktreeTool) IsReadOnly(input types.ToolInput) bool { return false }

func (t *EnterWorktreeTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *EnterWorktreeTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *EnterWorktreeTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}

func (t *ExitWorktreeTool) Name() string { return "worktree_exit" }

func (t *ExitWorktreeTool) Description() string {
	return "Exits an active local worktree session, optionally removing worktree files."
}

func (t *ExitWorktreeTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"action":          {Type: "string", Description: "keep or remove", Enum: []string{"keep", "remove"}},
			"discard_changes": {Type: "boolean", Description: "Required true when action=remove to acknowledge deletion."},
		},
	}
}

func (t *ExitWorktreeTool) Execute(_ context.Context, input types.ToolInput, _ types.ToolContext) (types.ToolResult, error) {
	var in exitWorktreeInput
	if err := json.Unmarshal(input, &in); err != nil {
		return exitWorktreeResult(exitWorktreeOutput{Success: false, Action: "keep", Message: "failed", Error: fmt.Sprintf("invalid input: %v", err)}, true)
	}
	action := strings.ToLower(strings.TrimSpace(in.Action))
	if action == "" {
		action = "keep"
	}
	if action != "keep" && action != "remove" {
		return exitWorktreeResult(exitWorktreeOutput{Success: false, Action: action, Message: "failed", Error: "action must be keep or remove"}, true)
	}

	orchestrationState.mu.Lock()
	defer orchestrationState.mu.Unlock()
	if orchestrationState.worktree == nil {
		return exitWorktreeResult(exitWorktreeOutput{Success: false, Action: action, Message: "failed", Error: "no active worktree session"}, true)
	}
	current := *orchestrationState.worktree

	removed := false
	if action == "remove" {
		if !in.DiscardChanges {
			return exitWorktreeResult(exitWorktreeOutput{Success: false, Action: action, WorktreeName: current.Name, WorktreePath: current.Path, Message: "failed", Error: "discard_changes must be true when action=remove"}, true)
		}
		if err := os.RemoveAll(current.Path); err != nil {
			return exitWorktreeResult(exitWorktreeOutput{Success: false, Action: action, WorktreeName: current.Name, WorktreePath: current.Path, Message: "failed", Error: err.Error()}, true)
		}
		removed = true
	}
	orchestrationState.worktree = nil
	return exitWorktreeResult(exitWorktreeOutput{Success: true, Action: action, WorktreeName: current.Name, WorktreePath: current.Path, Removed: removed, Message: "worktree session cleared"}, false)
}

func exitWorktreeResult(out exitWorktreeOutput, isError bool) (types.ToolResult, error) {
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b), IsError: isError}, nil
}

func (t *ExitWorktreeTool) IsReadOnly(input types.ToolInput) bool { return false }

func (t *ExitWorktreeTool) IsDestructive(input types.ToolInput) bool {
	var in exitWorktreeInput
	if err := json.Unmarshal(input, &in); err != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(in.Action), "remove")
}

func (t *ExitWorktreeTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *ExitWorktreeTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}
