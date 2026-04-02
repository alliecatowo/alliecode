package tools

import (
	"strings"

	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/types"
)

func permissionPolicyRules(family toolFamily, toolName string, toolCtx types.ToolContext) []permissions.Rule {
	rules := make([]permissions.Rule, 0, 2)
	if rule, ok := familyReadOnlyAllowRule(family, toolName); ok {
		rules = append(rules, rule)
	}
	if rule, ok := familyAutoMutationAskRule(family, toolName, toolCtx); ok {
		rules = append(rules, rule)
	}
	return rules
}

func familyReadOnlyAllowRule(family toolFamily, toolName string) (permissions.Rule, bool) {
	if toolName == "" {
		return permissions.Rule{}, false
	}
	if !permissionFamilySupportsReadOnlyPolicy(family) {
		return permissions.Rule{}, false
	}
	if !permissionFamilyReadOnlyToolNames[toolName] {
		return permissions.Rule{}, false
	}
	return permissions.Rule{Source: permissions.RuleSourcePolicy, Tool: toolName, Decision: permissions.DecisionAllow}, true
}

func familyAutoMutationAskRule(family toolFamily, toolName string, toolCtx types.ToolContext) (permissions.Rule, bool) {
	if toolName == "" || !toolCtx.IsNonInteractive {
		return permissions.Rule{}, false
	}
	if !permissionFamilySupportsMutationPolicy(family) {
		return permissions.Rule{}, false
	}
	if permissionFamilyReadOnlyToolNames[toolName] {
		return permissions.Rule{}, false
	}
	if strings.HasPrefix(toolName, "mcp__") {
		return permissions.Rule{Source: permissions.RuleSourcePolicy, Tool: toolName, Decision: permissions.DecisionAsk}, true
	}
	return permissions.Rule{Source: permissions.RuleSourcePolicy, Tool: toolName, Decision: permissions.DecisionAsk}, true
}

var permissionFamilyReadOnlyToolNames = map[string]bool{
	"read":              true,
	"ls":                true,
	"glob":              true,
	"grep":              true,
	"webfetch":          true,
	"websearch":         true,
	"task_get":          true,
	"task_list":         true,
	"task_output":       true,
	"team_list":         true,
	"team_status":       true,
	"mcp_resource_list": true,
	"mcp_resource_read": true,
	"mcp_auth_status":   true,
}

func permissionFamilySupportsReadOnlyPolicy(family toolFamily) bool {
	switch family {
	case toolFamilyFile, toolFamilyWeb, toolFamilyTask, toolFamilyTeam, toolFamilyMCP:
		return true
	default:
		return false
	}
}

func permissionFamilySupportsMutationPolicy(family toolFamily) bool {
	switch family {
	case toolFamilyShell, toolFamilyFile, toolFamilyWeb, toolFamilyTask, toolFamilyTeam, toolFamilyMCP:
		return true
	default:
		return false
	}
}
