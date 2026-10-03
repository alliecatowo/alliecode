package commands

import (
	"github.com/alliecatowo/alliecode/internal/permissions"
)

func stageGState() *RuntimeState {
	return &RuntimeState{
		PermissionMode:     permissions.ModeDefault,
		ConfigValues:       map[string]string{"settings.output-style": "human", "settings.output-format": "text", "settings.transport": "local"},
		MCPConnections:     map[string]bool{"local": true},
		SandboxMode:        "workspace-write",
		RemoteEnvironment:  "default",
		ProjectPaths:       []string{"."},
		PluginMarketplaces: []string{"https://plugins.example.dev"},
	}
}
