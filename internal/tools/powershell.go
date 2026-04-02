package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/types"
)

// PowerShellTool is a safe local shell variant that mirrors Bash safeguards.
type PowerShellTool struct{}

type powerShellInput struct {
	Command     string `json:"command"`
	Description string `json:"description"`
	Timeout     int    `json:"timeout"`
}

func (t *PowerShellTool) Name() string { return "powershell" }

func (t *PowerShellTool) Description() string {
	return "Executes a local shell command with PowerShell-compatible naming and Bash safety guardrails."
}

func (t *PowerShellTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"command": {
				Type:        "string",
				Description: "The command to execute.",
			},
			"description": {
				Type:        "string",
				Description: "A brief description of what the command does.",
			},
			"timeout": {
				Type:        "integer",
				Description: "Timeout in milliseconds. Defaults to 120000 (2 minutes).",
			},
		},
		Required: []string{"command"},
	}
}

func (t *PowerShellTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	started := time.Now()
	var in powerShellInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}

	command := strings.TrimSpace(in.Command)
	if command == "" {
		return types.ToolResult{Content: "command is required", IsError: true}, nil
	}

	if denyMsg := runShellPreflight(command, toolCtx.WorkingDir, nil, false); denyMsg != "" {
		return types.ToolResult{Content: denyMsg, IsError: true}, nil
	}
	preflight := evaluateBashPreflight(command, toolCtx.WorkingDir, nil)

	timeout := 120 * time.Second
	if in.Timeout > 0 {
		timeout = time.Duration(in.Timeout) * time.Millisecond
	}

	provider := selectShellProvider(t.Name())
	res, err := executeWithShellProvider(ctx, toolCtx, provider, command, timeout)
	if err != nil {
		return res, err
	}
	classify := permissions.ClassifyCommandDetailed(command)
	riskCodes := make([]string, 0, len(classify.Reasons))
	for _, reason := range classify.Reasons {
		riskCodes = append(riskCodes, reason.Code)
	}
	audit := &shellExecutionAudit{
		RiskLevel:      classify.Level.String(),
		RiskCodes:      riskCodes,
		PolicyDecision: preflight.Behavior,
		PolicyRuleID:   preflight.RuleID,
		PolicyReason:   preflight.Reason,
	}
	meta := formatShellExecutionBlock(t.Name(), provider.Type(), toolCtx.WorkingDir, timeout, time.Since(started), command, res.Content, res.IsError, audit)
	if strings.TrimSpace(res.Content) == "" {
		res.Content = meta
	} else {
		res.Content = res.Content + "\n\n" + meta
	}
	return res, nil
}

func (t *PowerShellTool) IsReadOnly(input types.ToolInput) bool { return false }

func (t *PowerShellTool) IsDestructive(input types.ToolInput) bool {
	var in powerShellInput
	if err := json.Unmarshal(input, &in); err != nil {
		return true
	}
	command := strings.TrimSpace(in.Command)
	return len(classifyCommand(command)) > 0
}

func (t *PowerShellTool) IsConcurrencySafe(input types.ToolInput) bool { return false }

func (t *PowerShellTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	var in powerShellInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.PermissionDenied
	}
	command := strings.TrimSpace(in.Command)
	if command == "" {
		return types.PermissionDenied
	}
	baseDecision := evaluateToolPermissionByFamily(toolFamilyShell, "powershell", input, toolCtx, types.PermissionAsk)
	if validateBashCommand(command) != "" {
		return types.PermissionDenied
	}
	if len(classifyCommand(command)) > 0 {
		return maxPermissionSeverity(baseDecision, types.PermissionAsk)
	}
	return baseDecision
}
