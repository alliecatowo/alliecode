package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

// REPLTool relays a command to a local primitive tool.
type REPLTool struct{}

type replInput struct {
	ToolName string          `json:"tool_name"`
	Input    json.RawMessage `json:"input"`
}

type replOutput struct {
	Tool    string `json:"tool"`
	IsError bool   `json:"is_error"`
	Result  string `json:"result"`
}

var replAllowedTargets = map[string]struct{}{
	"bash":         {},
	"read":         {},
	"write":        {},
	"edit":         {},
	"glob":         {},
	"grep":         {},
	"notebookedit": {},
	"agent":        {},
	"powershell":   {},
}

func (t *REPLTool) Name() string { return "repl" }

func (t *REPLTool) Description() string {
	return "Relays tool input to a local primitive tool and returns the wrapped result."
}

func (t *REPLTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"tool_name": {
				Type:        "string",
				Description: "Target primitive tool name.",
			},
			"input": {
				Type:        "object",
				Description: "JSON input forwarded to the target tool.",
			},
		},
		Required: []string{"tool_name", "input"},
	}
}

func (t *REPLTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	in, target, targetInput, errResult := resolveReplTarget(input)
	if errResult != nil {
		return *errResult, nil
	}

	res, err := target.Execute(ctx, targetInput, toolCtx)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("relay execution error: %v", err), IsError: true}, nil
	}

	out := replOutput{Tool: in.ToolName, IsError: res.IsError, Result: res.Content}
	b, err := json.Marshal(out)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b), IsError: res.IsError}, nil
}

func (t *REPLTool) IsReadOnly(input types.ToolInput) bool {
	_, target, targetInput, errResult := resolveReplTarget(input)
	if errResult != nil {
		return false
	}
	return target.IsReadOnly(targetInput)
}

func (t *REPLTool) IsDestructive(input types.ToolInput) bool {
	_, target, targetInput, errResult := resolveReplTarget(input)
	if errResult != nil {
		return true
	}
	return target.IsDestructive(targetInput)
}

func (t *REPLTool) IsConcurrencySafe(input types.ToolInput) bool {
	_, target, targetInput, errResult := resolveReplTarget(input)
	if errResult != nil {
		return false
	}
	return target.IsConcurrencySafe(targetInput)
}

func (t *REPLTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	_, target, targetInput, errResult := resolveReplTarget(input)
	if errResult != nil {
		return types.PermissionDenied
	}
	return target.CheckPermissions(targetInput, toolCtx)
}

func resolveReplTarget(input types.ToolInput) (replInput, types.Tool, types.ToolInput, *types.ToolResult) {
	var in replInput
	if err := json.Unmarshal(input, &in); err != nil {
		res := types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}
		return replInput{}, nil, nil, &res
	}
	in.ToolName = strings.TrimSpace(in.ToolName)
	if in.ToolName == "" {
		res := types.ToolResult{Content: "tool_name is required", IsError: true}
		return replInput{}, nil, nil, &res
	}
	normalized := strings.ToLower(in.ToolName)
	if normalized == "repl" {
		res := types.ToolResult{Content: "repl cannot relay to itself", IsError: true}
		return replInput{}, nil, nil, &res
	}
	if _, ok := replAllowedTargets[normalized]; !ok {
		res := types.ToolResult{Content: fmt.Sprintf("tool %q is not available in repl relay", in.ToolName), IsError: true}
		return replInput{}, nil, nil, &res
	}

	if len(in.Input) == 0 {
		in.Input = json.RawMessage(`{}`)
	}

	target := findToolCaseInsensitive(DefaultRegistry(), in.ToolName)
	if target == nil {
		res := types.ToolResult{Content: fmt.Sprintf("tool %q is not registered", in.ToolName), IsError: true}
		return replInput{}, nil, nil, &res
	}

	return in, target, types.ToolInput(in.Input), nil
}

func findToolCaseInsensitive(r *Registry, name string) types.Tool {
	for _, tool := range r.All() {
		if strings.EqualFold(tool.Name(), name) {
			return tool
		}
	}
	return nil
}
