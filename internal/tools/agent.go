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

// AgentTool spawns a subagent to handle complex tasks.
type AgentTool struct{}

type agentInput struct {
	Prompt      string `json:"prompt"`
	Description string `json:"description"`
	Model       string `json:"model"`
	TaskID      string `json:"task_id"`
	TeamName    string `json:"team_name"`
	Name        string `json:"name"`
	Wait        bool   `json:"wait"`
	PollMs      int    `json:"poll_ms"`
}

// AgentRunner is a function that runs a subagent and returns its text response.
type AgentRunner func(ctx context.Context, prompt, model string, toolCtx types.ToolContext) (string, error)

// agentRunner holds the configured runner function, set via SetAgentRunner.
var agentRunner AgentRunner
var agentTaskManager *tasks.Manager

// SetAgentRunner configures the function used to spawn subagents.
// This should be called by the agent package during initialization.
func SetAgentRunner(r AgentRunner) {
	agentRunner = r
	agentTaskManager = tasks.NewManager(func(ctx context.Context, prompt, model string, toolCtx types.ToolContext) (string, error) {
		return r(ctx, prompt, model, toolCtx)
	})
}

func (t *AgentTool) Name() string { return "Agent" }

func (t *AgentTool) Description() string {
	return "Spawns a subagent to handle complex, multi-step tasks. The subagent runs independently with its own conversation context."
}

func (t *AgentTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"prompt": {
				Type:        "string",
				Description: "The task description for the subagent.",
			},
			"description": {
				Type:        "string",
				Description: "A brief description of what the subagent will do.",
			},
			"model": {
				Type:        "string",
				Description: "Optional model override for the subagent.",
			},
			"task_id": {
				Type:        "string",
				Description: "Optional existing task id to poll/resume.",
			},
			"team_name": {
				Type:        "string",
				Description: "Optional team lifecycle context name.",
			},
			"name": {
				Type:        "string",
				Description: "Optional teammate name for lifecycle tracking.",
			},
			"wait": {
				Type:        "boolean",
				Description: "If true, waits for completion before returning.",
			},
			"poll_ms": {
				Type:        "integer",
				Description: "Polling interval in milliseconds for wait/resume.",
			},
		},
		Required: []string{},
	}
}

func (t *AgentTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in agentInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}

	if agentTaskManager == nil {
		if agentRunner != nil {
			agentTaskManager = tasks.NewManager(func(ctx context.Context, prompt, model string, toolCtx types.ToolContext) (string, error) {
				return agentRunner(ctx, prompt, model, toolCtx)
			})
		}
	}
	if agentTaskManager == nil {
		return types.ToolResult{
			Content: "Subagent execution is not configured. The agent runner must be set up during initialization.",
			IsError: true,
		}, nil
	}

	pollInterval := 200 * time.Millisecond
	if in.PollMs > 0 {
		pollInterval = time.Duration(in.PollMs) * time.Millisecond
	}

	if in.TaskID != "" {
		if in.Wait {
			task, err := agentTaskManager.Wait(ctx, in.TaskID, pollInterval)
			if err != nil {
				return types.ToolResult{Content: fmt.Sprintf("Subagent resume error: %v", err), IsError: true}, nil
			}
			return types.ToolResult{Content: formatTaskResponse(task), IsError: task.Status == tasks.StatusFailed}, nil
		}
		task, ok := agentTaskManager.Get(in.TaskID)
		if !ok {
			return types.ToolResult{Content: fmt.Sprintf("task %q not found", in.TaskID), IsError: true}, nil
		}
		return types.ToolResult{Content: formatTaskResponse(task), IsError: task.Status == tasks.StatusFailed}, nil
	}

	if in.Prompt == "" {
		return types.ToolResult{Content: "prompt is required when task_id is not provided", IsError: true}, nil
	}

	teamName := strings.TrimSpace(in.TeamName)
	agentName := strings.TrimSpace(in.Name)
	if teamName != "" && agentName != "" {
		orchestrationState.mu.Lock()
		if team, ok := orchestrationState.teamStore.Get(teamName); ok {
			hasMember := false
			for _, m := range team.Members {
				if m.Name == agentName {
					hasMember = true
					break
				}
			}
			if !hasMember {
				updatedMembers := append([]teamMember(nil), team.Members...)
				updatedMembers = append(updatedMembers, teamMember{Name: agentName, Role: "member", JoinedAt: time.Now().UTC()})
				_, _ = orchestrationState.teamStore.Update(teamName, tasks.TeamUpdateParams{Members: updatedMembers})
			}
		}
		orchestrationState.mu.Unlock()
	}

	task, err := agentTaskManager.Start(ctx, in.Prompt, in.Model, toolCtx)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Subagent start error: %v", err), IsError: true}, nil
	}

	if teamName != "" {
		if agentName == "" {
			agentName = task.ID
		}
		_, _ = agentTaskManager.RecordMessage(task.ID, "team_lead", agentName, "spawn", in.Description)
		orchestrationState.mu.Lock()
		_, _ = orchestrationState.teamStore.AddMessage(teamName, "team_lead", agentName, "spawn", in.Prompt)
		orchestrationState.mu.Unlock()
	}

	if in.Wait {
		task, err = agentTaskManager.Wait(ctx, task.ID, pollInterval)
		if err != nil {
			return types.ToolResult{Content: fmt.Sprintf("Subagent wait error: %v", err), IsError: true}, nil
		}
	}

	return types.ToolResult{Content: formatTaskResponse(task), IsError: task.Status == tasks.StatusFailed}, nil
}

func formatTaskResponse(task tasks.Task) string {
	if task.Status == tasks.StatusCompleted {
		return fmt.Sprintf("task_id: %s\nstatus: %s\nresult:\n%s", task.ID, task.Status, task.Result)
	}
	if task.Status == tasks.StatusFailed {
		return fmt.Sprintf("task_id: %s\nstatus: %s\nerror: %s", task.ID, task.Status, task.Error)
	}
	return fmt.Sprintf("task_id: %s\nstatus: %s", task.ID, task.Status)
}

func (t *AgentTool) IsReadOnly(input types.ToolInput) bool {
	return false
}

func (t *AgentTool) IsDestructive(input types.ToolInput) bool {
	return false
}

func (t *AgentTool) IsConcurrencySafe(input types.ToolInput) bool {
	return true
}

func (t *AgentTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}
