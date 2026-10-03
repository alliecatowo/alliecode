package commands

import "strings"

type correctiveLoop struct {
	Area   string
	State  string
	Action string
	Next   string
}

func correctiveLoopsForState(state *RuntimeState) []correctiveLoop {
	if state == nil {
		return []correctiveLoop{
			{Area: "provider", State: "missing", Action: "/provider set ollama", Next: "/provider doctor"},
			{Area: "model", State: "missing", Action: "/model ollama/llama3", Next: "/model doctor"},
			{Area: "permissions", State: "unknown", Action: "/permissions status", Next: "/permissions set auto"},
			{Area: "settings", State: "unknown", Action: "/config doctor", Next: "/config repair"},
			{Area: "history", State: "unknown", Action: "/history status", Next: "/history latest"},
			{Area: "session", State: "unknown", Action: "/session status", Next: "/session diagnostics"},
			{Area: "mcp", State: "unknown", Action: "/mcp diagnostics", Next: "/mcp repair auto"},
		}
	}
	selection := RuntimeSelectionTruth(state)

	providerState := "ready"
	providerAction := "/provider status"
	providerNext := "/provider doctor"
	if strings.TrimSpace(selection.ProviderName) == "" {
		providerState = "missing"
		providerAction = "/provider set ollama"
	} else if !selection.ProviderReady {
		providerState = "degraded"
		providerAction = "/provider doctor"
		providerNext = "/login provider " + strings.TrimSpace(selection.ProviderName)
	}

	modelState := "ready"
	modelAction := "/model doctor"
	modelNext := "/status"
	if strings.TrimSpace(selection.ModelName) == "" {
		modelState = "missing"
		provider := strings.TrimSpace(selection.ProviderName)
		if provider == "" {
			provider = "ollama"
		}
		modelAction = "/model " + provider + "/<model>"
		modelNext = "/model doctor"
	}

	permissionState := "ready"
	permissionAction := "/permissions summary"
	permissionNext := "/permissions denials"
	if strings.EqualFold(strings.TrimSpace(modeString(state.PermissionMode)), "bypass") {
		permissionState = "risky"
		permissionAction = "/permissions set auto"
	}

	settingsState := "ready"
	settingsAction := "/config status"
	settingsNext := "/config repair"
	if state.ConfigValues == nil || strings.TrimSpace(state.ConfigValues["settings.output-style"]) == "" || strings.TrimSpace(state.ConfigValues["settings.output-format"]) == "" || strings.TrimSpace(state.ConfigValues["settings.transport"]) == "" {
		settingsState = "degraded"
		settingsAction = "/config doctor"
	}

	historyState := "ready"
	historyAction := "/history status"
	historyNext := "/history latest"
	if len(state.HistoryEntries) == 0 {
		historyState = "missing"
		historyAction = "/history list"
	}

	sessionState := "ready"
	sessionAction := "/session status"
	sessionNext := "/session diagnostics"
	if strings.EqualFold(strings.TrimSpace(state.TransportMode), "remote") && strings.TrimSpace(state.RemoteSessionURL) == "" {
		sessionState = "degraded"
		sessionAction = "/session set-url <url>"
	}
	if strings.EqualFold(strings.TrimSpace(state.SessionMode), "client") && !state.SessionConnected {
		sessionState = "degraded"
		sessionAction = "/session connect <addr> <token>"
		sessionNext = "/session token show"
	}

	mcpState := "ready"
	mcpAction := "/mcp status"
	mcpNext := "/mcp list-tools"
	if len(state.MCPConnections) == 0 {
		mcpState = "missing"
		mcpAction = "/mcp add <name> stdio <command>"
		mcpNext = "/mcp doctor"
	} else {
		connected := 0
		for _, ok := range state.MCPConnections {
			if ok {
				connected++
			}
		}
		if connected == 0 {
			mcpState = "degraded"
			mcpAction = "/mcp connect <name>"
			mcpNext = "/mcp diagnostics"
		}
	}

	return []correctiveLoop{
		{Area: "provider", State: providerState, Action: providerAction, Next: providerNext},
		{Area: "model", State: modelState, Action: modelAction, Next: modelNext},
		{Area: "permissions", State: permissionState, Action: permissionAction, Next: permissionNext},
		{Area: "settings", State: settingsState, Action: settingsAction, Next: settingsNext},
		{Area: "history", State: historyState, Action: historyAction, Next: historyNext},
		{Area: "session", State: sessionState, Action: sessionAction, Next: sessionNext},
		{Area: "mcp", State: mcpState, Action: mcpAction, Next: mcpNext},
	}
}
