package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/alliecatowo/alliecode/internal/types"
)

type planModeState struct {
	mu      sync.Mutex
	enabled bool
}

var sessionPlanMode = &planModeState{}

func resetPlanModeStateForTests() {
	sessionPlanMode.mu.Lock()
	defer sessionPlanMode.mu.Unlock()
	sessionPlanMode.enabled = false
}

// EnterPlanModeTool marks the local session as being in plan mode.
type EnterPlanModeTool struct{}

func (t *EnterPlanModeTool) Name() string { return "enter_plan_mode" }

func (t *EnterPlanModeTool) Description() string {
	return "Enters plan mode so the model can explore and design before implementation."
}

func (t *EnterPlanModeTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type:       "object",
		Properties: map[string]types.PropertySchema{},
	}
}

func (t *EnterPlanModeTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	if err := requireEmptyObjectInput(input); err != nil {
		return types.ToolResult{Content: err.Error(), IsError: true}, nil
	}

	sessionPlanMode.mu.Lock()
	sessionPlanMode.enabled = true
	sessionPlanMode.mu.Unlock()

	out := struct {
		Message string `json:"message"`
		Mode    string `json:"mode"`
	}{
		Message: "Entered plan mode. Focus on exploration and implementation design.",
		Mode:    "plan",
	}
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b)}, nil
}

func (t *EnterPlanModeTool) IsReadOnly(input types.ToolInput) bool { return true }

func (t *EnterPlanModeTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *EnterPlanModeTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *EnterPlanModeTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}

// ExitPlanModeTool leaves plan mode and returns a coding-ready confirmation.
type ExitPlanModeTool struct{}

type exitPlanModePrompt struct {
	Tool   string `json:"tool"`
	Prompt string `json:"prompt"`
}

type exitPlanModeInput struct {
	AllowedPrompts []exitPlanModePrompt `json:"allowedPrompts"`
}

func (t *ExitPlanModeTool) Name() string { return "exit_plan_mode" }

func (t *ExitPlanModeTool) Description() string {
	return "Exits plan mode and confirms implementation can begin."
}

func (t *ExitPlanModeTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"allowedPrompts": {
				Type:        "array",
				Description: "Optional prompt-based permissions requested by the plan.",
				Items:       &types.PropertySchema{Type: "object"},
			},
		},
	}
}

func (t *ExitPlanModeTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in exitPlanModeInput
	if len(input) > 0 {
		if err := json.Unmarshal(input, &in); err != nil {
			return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
		}
	}
	for i, prompt := range in.AllowedPrompts {
		if strings.TrimSpace(prompt.Tool) == "" {
			return types.ToolResult{Content: fmt.Sprintf("allowedPrompts[%d].tool is required", i), IsError: true}, nil
		}
		if strings.TrimSpace(prompt.Prompt) == "" {
			return types.ToolResult{Content: fmt.Sprintf("allowedPrompts[%d].prompt is required", i), IsError: true}, nil
		}
	}

	sessionPlanMode.mu.Lock()
	wasEnabled := sessionPlanMode.enabled
	if wasEnabled {
		sessionPlanMode.enabled = false
	}
	sessionPlanMode.mu.Unlock()

	if !wasEnabled {
		return types.ToolResult{Content: "You are not in plan mode. Use enter_plan_mode before exiting plan mode.", IsError: true}, nil
	}

	out := struct {
		Message             string `json:"message"`
		Mode                string `json:"mode"`
		AllowedPromptsCount int    `json:"allowed_prompts_count"`
	}{
		Message:             "Exited plan mode. You can now start implementation.",
		Mode:                "default",
		AllowedPromptsCount: len(in.AllowedPrompts),
	}
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b)}, nil
}

func (t *ExitPlanModeTool) IsReadOnly(input types.ToolInput) bool { return true }

func (t *ExitPlanModeTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *ExitPlanModeTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *ExitPlanModeTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}

func requireEmptyObjectInput(input types.ToolInput) error {
	if len(input) == 0 {
		return nil
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(input, &obj); err != nil {
		return fmt.Errorf("invalid input: %v", err)
	}
	if len(obj) > 0 {
		return fmt.Errorf("this tool accepts no input fields")
	}
	return nil
}
