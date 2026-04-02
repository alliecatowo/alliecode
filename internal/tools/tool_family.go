package tools

type toolFamily string

const (
	toolFamilyGeneric toolFamily = "generic"
	toolFamilyFile    toolFamily = "file"
	toolFamilyShell   toolFamily = "shell"
	toolFamilyWeb     toolFamily = "web"
	toolFamilyMCP     toolFamily = "mcp"
	toolFamilyTask    toolFamily = "task"
	toolFamilyTeam    toolFamily = "team"
)
