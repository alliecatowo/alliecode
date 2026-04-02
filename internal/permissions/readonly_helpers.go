package permissions

import "strings"

var readOnlyToolNames = map[string]bool{
	"read":              true,
	"fileread":          true,
	"glob":              true,
	"grep":              true,
	"search":            true,
	"ls":                true,
	"cat":               true,
	"task_get":          true,
	"task_list":         true,
	"task_output":       true,
	"team_list":         true,
	"team_status":       true,
	"websearch":         true,
	"tool_search":       true,
	"mcp_resource_list": true,
	"mcp_resource_read": true,
	"mcp_auth_status":   true,
}

func isReadOnlyToolName(name string) bool {
	return readOnlyToolNames[strings.ToLower(name)]
}
