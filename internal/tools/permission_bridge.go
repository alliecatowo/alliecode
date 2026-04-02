package tools

import (
	"strings"

	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/types"
)

func evaluateToolPermission(toolName string, input types.ToolInput, toolCtx types.ToolContext, fallback types.ToolPermission) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyGeneric, toolName, input, toolCtx, fallback)
}

func evaluateToolPermissionByFamily(family toolFamily, toolName string, input types.ToolInput, toolCtx types.ToolContext, fallback types.ToolPermission) types.ToolPermission {
	mode := permissionModeForToolContext(toolCtx)
	canonical := canonicalPermissionToolName(toolName)
	rules := permissionPolicyRules(family, canonical, toolCtx)
	engine := permissions.NewEngine(mode, rules)
	decision := engine.Check(canonical, input)
	mapped := fallback
	switch decision {
	case permissions.DecisionAllow:
		mapped = types.PermissionAllowed
	case permissions.DecisionDeny:
		mapped = types.PermissionDenied
	case permissions.DecisionAsk:
		mapped = types.PermissionAsk
	default:
		mapped = fallback
	}
	return tightenPermissionForHighRisk(canonical, input, toolCtx, mapped)
}

func permissionModeForToolContext(toolCtx types.ToolContext) permissions.Mode {
	if toolCtx.IsNonInteractive {
		return permissions.ModeAuto
	}
	return permissions.ModeDefault
}

func canonicalPermissionToolName(name string) string {
	normalized := strings.TrimSpace(strings.ToLower(name))
	if normalized == "" {
		return normalized
	}
	if mapped, ok := permissionToolAliases[normalized]; ok {
		return mapped
	}
	return normalized
}

var permissionToolAliases = map[string]string{
	"read":              "read",
	"fileread":          "read",
	"ls":                "ls",
	"glob":              "glob",
	"grep":              "grep",
	"write":             "write",
	"filewrite":         "write",
	"edit":              "edit",
	"fileedit":          "edit",
	"bash":              "bash",
	"shell":             "bash",
	"powershell":        "powershell",
	"webfetch":          "webfetch",
	"websearch":         "websearch",
	"task_get":          "task_get",
	"task_list":         "task_list",
	"task_create":       "task_create",
	"task_update":       "task_update",
	"task_stop":         "task_stop",
	"task_output":       "task_output",
	"team_create":       "team_create",
	"team_list":         "team_list",
	"team_status":       "team_status",
	"team_update":       "team_update",
	"team_delete":       "team_delete",
	"send_message":      "send_message",
	"mcp_tool_invoke":   "mcp_tool_invoke",
	"mcp_resource_list": "mcp_resource_list",
	"mcp_resource_read": "mcp_resource_read",
	"mcp_auth_local":    "mcp_auth_local",
	"mcp_auth_status":   "mcp_auth_status",
}
