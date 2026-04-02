package tools

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type shellExecutionAudit struct {
	RiskLevel      string
	RiskCodes      []string
	PolicyDecision string
	PolicyRuleID   string
	PolicyReason   string
}

func formatCommandTimeout(timeout time.Duration, output string) string {
	return fmt.Sprintf("Command timed out after %v\n%s", timeout, output)
}

func formatCommandExit(output string, err error) string {
	return fmt.Sprintf("%s\nExit code: %v", output, err)
}

func truncateCommandOutput(output string, maxLines, maxChars int) string {
	if maxLines > 0 {
		lines := strings.Split(output, "\n")
		if len(lines) > maxLines {
			lines = append(lines[:maxLines], fmt.Sprintf("... (truncated to %d lines)", maxLines))
			output = strings.Join(lines, "\n")
		}
	}

	if maxChars > 0 && len(output) > maxChars {
		output = output[:maxChars] + "\n... (truncated)"
	}

	return output
}

func formatShellExecutionBlock(toolName, providerName, workingDir string, timeout time.Duration, duration time.Duration, command string, output string, isError bool, audit *shellExecutionAudit) string {
	status := "ok"
	if isError {
		status = "error"
	}
	fields := []structuredField{
		{Key: "tool", Value: strings.ToLower(strings.TrimSpace(toolName))},
		{Key: "provider", Value: providerName},
		{Key: "status", Value: status},
		{Key: "timeout_ms", Value: strconv.FormatInt(timeout.Milliseconds(), 10)},
		{Key: "duration_ms", Value: strconv.FormatInt(duration.Milliseconds(), 10)},
		{Key: "working_dir", Value: workingDir},
		{Key: "command_len", Value: strconv.Itoa(len(command))},
		{Key: "output_bytes", Value: strconv.Itoa(len(output))},
	}
	if audit != nil {
		fields = append(fields,
			structuredField{Key: "risk_level", Value: audit.RiskLevel},
			structuredField{Key: "risk_codes", Value: strings.Join(audit.RiskCodes, ",")},
			structuredField{Key: "policy_decision", Value: audit.PolicyDecision},
			structuredField{Key: "policy_rule_id", Value: audit.PolicyRuleID},
			structuredField{Key: "policy_reason", Value: audit.PolicyReason},
		)
	}
	return renderStructuredBlock("shell_execution", fields)
}
