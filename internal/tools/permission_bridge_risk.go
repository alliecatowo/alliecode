package tools

import (
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/types"
)

func tightenPermissionForHighRisk(canonical string, input types.ToolInput, toolCtx types.ToolContext, base types.ToolPermission) types.ToolPermission {
	decision := base
	if canonical == "bash" || canonical == "powershell" {
		var payload struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal(input, &payload); err == nil {
			classification := permissions.ClassifyCommandDetailed(payload.Command)
			for _, reason := range classification.Reasons {
				switch reason.Code {
				case "comment_quote_desync", "escaped_newline_operator":
					decision = maxPermissionSeverity(decision, types.PermissionDenied)
				case "escaped_operator", "suspicious_substitution_combo":
					decision = maxPermissionSeverity(decision, types.PermissionAsk)
				}
			}
			switch {
			case classification.Level >= permissions.RiskCritical:
				decision = maxPermissionSeverity(decision, types.PermissionDenied)
			case classification.Level >= permissions.RiskHigh:
				if toolCtx.IsNonInteractive {
					decision = maxPermissionSeverity(decision, types.PermissionDenied)
				} else {
					decision = maxPermissionSeverity(decision, types.PermissionAsk)
				}
			}
		}
	}

	if canonical == "task_update" || canonical == "task_stop" || canonical == "team_delete" {
		if toolCtx.IsNonInteractive {
			decision = maxPermissionSeverity(decision, types.PermissionAsk)
		}
	}

	if canonical == "mcp_auth_local" {
		if toolCtx.IsNonInteractive {
			decision = maxPermissionSeverity(decision, types.PermissionAsk)
		}
	}

	if canonical == "mcp_tool_invoke" {
		var payload struct {
			ToolName string `json:"tool_name"`
		}
		if err := json.Unmarshal(input, &payload); err == nil && classifyMutationIntent(payload.ToolName) {
			decision = maxPermissionSeverity(decision, types.PermissionAsk)
		}
	}

	if canonical == "write" || canonical == "edit" {
		path := extractFilePath(input)
		if path != "" {
			norm := strings.ToLower(filepath.ToSlash(path))
			if strings.Contains(norm, "/.ssh/") || strings.Contains(norm, "/.gnupg/") || strings.HasSuffix(norm, "/.env") || strings.Contains(norm, "/.env.") || strings.Contains(norm, "/etc/") || strings.Contains(norm, "/.claude/") || strings.HasSuffix(norm, "/.gitconfig") || strings.HasSuffix(norm, "/.gitmodules") || strings.HasSuffix(norm, "/.bashrc") || strings.HasSuffix(norm, "/.zshrc") {
				if toolCtx.IsNonInteractive {
					decision = maxPermissionSeverity(decision, types.PermissionDenied)
				} else {
					decision = maxPermissionSeverity(decision, types.PermissionAsk)
				}
			}
		}
		if contentLooksSensitive(input) {
			decision = maxPermissionSeverity(decision, types.PermissionAsk)
		}
	}

	return decision
}

func extractFilePath(input types.ToolInput) string {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(input, &payload); err != nil {
		return ""
	}
	raw, ok := payload["file_path"]
	if !ok {
		return ""
	}
	var path string
	if err := json.Unmarshal(raw, &path); err != nil {
		return ""
	}
	return strings.TrimSpace(path)
}

func contentLooksSensitive(input types.ToolInput) bool {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(input, &payload); err != nil {
		return false
	}
	raw, ok := payload["content"]
	if !ok {
		raw = payload["new_string"]
	}
	if len(raw) == 0 {
		return false
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return false
	}
	lower := strings.ToLower(text)
	return strings.Contains(text, "BEGIN PRIVATE KEY") || strings.Contains(lower, "aws_secret_access_key") || strings.Contains(lower, "authorization: bearer") || strings.Contains(lower, "token=")
}

func maxPermissionSeverity(a, b types.ToolPermission) types.ToolPermission {
	rank := func(v types.ToolPermission) int {
		switch v {
		case types.PermissionDenied:
			return 3
		case types.PermissionAsk:
			return 2
		case types.PermissionAllowed:
			return 1
		default:
			return 0
		}
	}
	if rank(b) > rank(a) {
		return b
	}
	return a
}

func classifyMutationIntent(name string) bool {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return false
	}
	for _, token := range []string{"write", "edit", "delete", "remove", "update", "create", "invoke", "auth"} {
		if strings.Contains(name, token) {
			return true
		}
	}
	return false
}
