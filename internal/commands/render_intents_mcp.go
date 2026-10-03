package commands

import (
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

func mcpListIntents(rows []types.RenderTableRow) []types.RenderIntent {
	return []types.RenderIntent{tableIntent("MCP servers", "Configured MCP server inventory.", []string{"Name", "Transport/Status", "Auth", "Connected"}, rows...)}
}

func mcpSettingsIntents(redirect bool) []types.RenderIntent {
	return []types.RenderIntent{
		detailRowsIntent("MCP settings", "Global MCP rendering and redirect preferences.", detailRow("Redirect", boolState(redirect, "yes", "no"), statusWord(redirect, "enabled", "disabled"), "Whether MCP auth flows should redirect externally.")),
	}
}

func mcpReconnectIntents(name, quickFix string) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("MCP reconnect", "Marked an MCP server connected and updated the recommended follow-up.", field("Server", defaultDash(name)), field("Connected", "yes")),
		actionListIntent("MCP reconnect actions", "Suggested follow-up commands for the reconnected server.", action("Inspect status", quickFix, "Verify connection and auth state for this server.", "recommended")),
	}
}

func mcpToggleIntents(mode, target string, changed int, message string) []types.RenderIntent {
	return []types.RenderIntent{
		detailRowsIntent("MCP toggle", "Bulk or server-scoped enable/disable result.",
			detailRow("Action", defaultDash(mode), "info", "Requested MCP toggle action."),
			detailRow("Target", defaultDash(target), "info", "Server name or all."),
			detailRow("Changed", fmt.Sprintf("%d", changed), statusWord(changed > 0, "updated", "noop"), defaultDash(message)),
		),
	}
}

func mcpDoctorIntents(manager bool, servers, connected, doctorRuns int, quickFix string) []types.RenderIntent {
	return []types.RenderIntent{diagnosticsIntent("MCP doctor", boolState(connected > 0, "connected", "disconnected"), "Server connectivity and readiness summary.", []types.RenderDiagnostic{
		diagnostic("Manager", boolState(manager, "ok", "warn"), boolState(manager, "available", "unavailable")),
		diagnostic("Servers", "info", fmt.Sprintf("%d", servers)),
		diagnostic("Connected", "info", fmt.Sprintf("%d", connected)),
		diagnostic("Doctor runs", "info", fmt.Sprintf("%d", doctorRuns)),
	}, []types.RenderActionHint{hint("Quick fix", quickFix)})}
}

func mcpDiagnosticsIntents(servers, connected, pending, failed, doctorRuns, repairRuns int, quickFix string) []types.RenderIntent {
	return []types.RenderIntent{diagnosticsIntent("MCP diagnostics", "diagnostics", "Connectivity and auth state across MCP servers.", []types.RenderDiagnostic{
		diagnostic("Servers", "info", fmt.Sprintf("%d", servers)),
		diagnostic("Connected", "info", fmt.Sprintf("%d", connected)),
		diagnostic("Pending", "info", fmt.Sprintf("%d", pending)),
		diagnostic("Failed", boolState(failed == 0, "ok", "warn"), fmt.Sprintf("%d", failed)),
		diagnostic("Doctor runs", "info", fmt.Sprintf("%d", doctorRuns)),
		diagnostic("Repair runs", "info", fmt.Sprintf("%d", repairRuns)),
	}, []types.RenderActionHint{hint("Quick fix", quickFix)})}
}

func mcpRepairIntents(mode string, changed bool, quickFix string, repairRuns int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("MCP repair", "Applied deterministic MCP repair flow.", field("Mode", mode), field("Changed", boolState(changed, "yes", "no")), field("Repair runs", fmt.Sprintf("%d", repairRuns))), actionHintsIntent("Actions", hint("Follow-up", quickFix))}
}

func mcpMutationIntents(title, name, transport string, changed bool) []types.RenderIntent {
	return []types.RenderIntent{
		detailRowsIntent(title, "Single MCP server mutation result.",
			detailRow("Server", defaultDash(name), "info", "Affected MCP server."),
			detailRow("Transport", defaultDash(transport), "info", "Configured transport."),
			detailRow("Changed", boolState(changed, "yes", "no"), statusWord(changed, "updated", "noop"), "Whether the MCP registry changed."),
		),
	}
}

func mcpConnectionIntents(title, name, status string, connected bool) []types.RenderIntent {
	return []types.RenderIntent{
		detailRowsIntent(title, "Connection result for a single MCP server.",
			detailRow("Server", defaultDash(name), "info", "Affected MCP server."),
			detailRow("Status", defaultDash(status), statusWord(connected, "connected", "disconnected"), "Reported connection state."),
		),
	}
}

func mcpStatusIntents(scope string, options []types.RenderOption) []types.RenderIntent {
	title := "MCP status"
	summary := "Connection and auth state for MCP servers."
	if strings.TrimSpace(scope) != "" && scope != "all" {
		title = "MCP server status"
		summary = "Connection and auth state for a single MCP server."
	}
	return []types.RenderIntent{optionListIntent(title, summary, options...)}
}

func mcpToolsIntents(serverFilter string, options []types.RenderOption) []types.RenderIntent {
	summary := "Registered MCP tools grouped by backing server."
	if strings.TrimSpace(serverFilter) != "" {
		summary = fmt.Sprintf("Registered MCP tools for %s.", serverFilter)
	}
	return []types.RenderIntent{optionListIntent("MCP tools", summary, options...)}
}

func mcpResourcesIntents(serverFilter string, options []types.RenderOption) []types.RenderIntent {
	summary := "Registered MCP resources across configured servers."
	if strings.TrimSpace(serverFilter) != "" {
		summary = fmt.Sprintf("Registered MCP resources for %s.", serverFilter)
	}
	return []types.RenderIntent{optionListIntent("MCP resources", summary, options...)}
}

func mcpAuthStatusIntents(options []types.RenderOption) []types.RenderIntent {
	return []types.RenderIntent{optionListIntent("MCP auth status", "Authentication state across MCP servers.", options...)}
}
