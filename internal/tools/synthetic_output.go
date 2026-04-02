package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alliecatowo/alliecode/internal/types"
)

// SyntheticOutputTool validates object input and returns it as structured payload.
type SyntheticOutputTool struct{}

type syntheticOutputResult struct {
	Data             string         `json:"data"`
	StructuredOutput map[string]any `json:"structured_output"`
}

func (t *SyntheticOutputTool) Name() string { return "synthetic_output" }

func (t *SyntheticOutputTool) Description() string {
	return "Returns structured JSON output in a deterministic passthrough envelope."
}

func (t *SyntheticOutputTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type:       "object",
		Properties: map[string]types.PropertySchema{},
	}
}

func (t *SyntheticOutputTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var payload map[string]any
	if err := json.Unmarshal(input, &payload); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("invalid input: %v", err), IsError: true}, nil
	}
	if payload == nil {
		payload = map[string]any{}
	}

	b, err := json.Marshal(syntheticOutputResult{
		Data:             "Structured output provided successfully",
		StructuredOutput: payload,
	})
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}

	return types.ToolResult{Content: string(b)}, nil
}

func (t *SyntheticOutputTool) IsReadOnly(input types.ToolInput) bool { return true }

func (t *SyntheticOutputTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *SyntheticOutputTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *SyntheticOutputTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}
