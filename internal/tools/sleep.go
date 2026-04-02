package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

// SleepTool pauses execution for a bounded duration.
type SleepTool struct{}

type sleepInput struct {
	DurationMs int `json:"duration_ms"`
}

type sleepResult struct {
	Status      string `json:"status"`
	RequestedMs int    `json:"requested_ms"`
	SleptMs     int    `json:"slept_ms"`
	Bounded     bool   `json:"bounded"`
}

const (
	minSleepDurationMs = 1
	maxSleepDurationMs = 600000
)

func (t *SleepTool) Name() string { return "Sleep" }

func (t *SleepTool) Description() string {
	return "Sleeps for a requested duration in milliseconds."
}

func (t *SleepTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"duration_ms": {
				Type:        "integer",
				Description: "Requested sleep duration in milliseconds; bounded to 1..600000.",
			},
		},
		Required: []string{"duration_ms"},
	}
}

func (t *SleepTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	var in sleepInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}

	requested := in.DurationMs
	bounded := requested
	if bounded < minSleepDurationMs {
		bounded = minSleepDurationMs
	}
	if bounded > maxSleepDurationMs {
		bounded = maxSleepDurationMs
	}
	result := sleepResult{
		Status:      "completed",
		RequestedMs: requested,
		SleptMs:     bounded,
		Bounded:     bounded != requested,
	}

	select {
	case <-ctx.Done():
		result.Status = "cancelled"
		b, _ := json.Marshal(result)
		return types.ToolResult{Content: string(b), IsError: true}, nil
	case <-time.After(time.Duration(bounded) * time.Millisecond):
	}

	b, err := json.Marshal(result)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("serialization error: %v", err), IsError: true}, nil
	}
	return types.ToolResult{Content: string(b)}, nil
}

func (t *SleepTool) IsReadOnly(input types.ToolInput) bool { return true }

func (t *SleepTool) IsDestructive(input types.ToolInput) bool { return false }

func (t *SleepTool) IsConcurrencySafe(input types.ToolInput) bool { return true }

func (t *SleepTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return types.PermissionAllowed
}
