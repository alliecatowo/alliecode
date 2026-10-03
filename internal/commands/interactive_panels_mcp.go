package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

func buildMCPInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	if state.MCPConnections == nil {
		state.MCPConnections = make(map[string]bool)
	}
	connected := 0
	names := make([]string, 0, len(state.MCPConnections))
	for name, ok := range state.MCPConnections {
		names = append(names, name)
		if ok {
			connected++
		}
	}
	sort.Strings(names)

	panel := InteractivePanel{
		Command:       "mcp",
		Title:         "mcp panel: /mcp",
		Subtitle:      fmt.Sprintf("servers=%d connected=%d", len(names), connected),
		HeaderIntents: mcpListIntents(panelMCPRows(names, state.MCPConnections)),
		Items: []InteractivePanelItem{
			{Key: "list", Section: "Overview", Label: "List servers", Detail: "Show configured MCP servers and connection states", Status: statusWord(len(names) > 0, "inventory", "empty"), ApplyInput: "/mcp list", ApplyMode: PanelApplySubmit, PreviewIntents: mcpListIntents(panelMCPRows(names, state.MCPConnections)), Preview: previewLines(fmt.Sprintf("Configured MCP servers: %d", len(names)))},
			{Key: "doctor", Section: "Diagnostics", Label: "MCP doctor", Detail: "Run manager/server diagnostics and quick-fix planning", Status: statusWord(connected > 0, "connected", "needs-setup"), ApplyInput: "/mcp doctor", ApplyMode: PanelApplySubmit, PreviewIntents: mcpDoctorIntents(false, len(names), connected, state.MCPDoctorCount, "/mcp add <name> stdio <command>"), Preview: previewLines(fmt.Sprintf("Connected servers: %d", connected), "Use doctor to validate MCP setup and suggested fixes.")},
			{Key: "diagnostics", Section: "Diagnostics", Label: "MCP diagnostics", Detail: "Render deterministic MCP diagnostics", Status: "diagnostics", ApplyInput: "/mcp diagnostics", ApplyMode: PanelApplySubmit, PreviewIntents: mcpDiagnosticsIntents(len(names), connected, 0, 0, state.MCPDoctorCount, state.MCPRepairCount, "/mcp doctor"), Preview: previewLines("/mcp diagnostics shows connectivity and repair-oriented metadata.")},
			{Key: "repair", Section: "Diagnostics", Label: "Repair MCP state", Detail: "Apply deterministic MCP repair flow", Status: "repair", ApplyInput: "/mcp repair", ApplyMode: PanelApplySubmit, PreviewIntents: mcpRepairIntents("auto", false, "/mcp doctor", state.MCPRepairCount), Preview: previewLines("/mcp repair chooses a deterministic repair mode based on current state.")},
		},
	}

	for _, name := range names {
		isConnected := state.MCPConnections[name]
		panel.Items = append(panel.Items,
			InteractivePanelItem{Key: name + ":status", Section: "Servers", Label: fmt.Sprintf("Server %s status", name), Detail: "Inspect this server only", Status: statusWord(isConnected, "connected", "disconnected"), ApplyInput: "/mcp status " + name, ApplyMode: PanelApplySubmit, PreviewIntents: mcpStatusIntents(name, []types.RenderOption{option(name, "local runtime connection inventory", statusWord(isConnected, "connected", "disconnected"), "/mcp auth-status "+name, true)}), Preview: previewLines(fmt.Sprintf("Server: %s", name), fmt.Sprintf("Connected: %t", isConnected))},
			InteractivePanelItem{Key: name + ":toggle", Section: "Servers", Label: fmt.Sprintf("%s %s", toggleVerb(isConnected, "Disconnect", "Connect"), name), Detail: "Toggle this server's connection state", Status: statusWord(isConnected, "live", "offline"), ApplyInput: toggleCommand(isConnected, "/mcp disconnect "+name, "/mcp connect "+name), ApplyMode: PanelApplySubmit, PreviewIntents: mcpConnectionIntents("MCP toggle target", name, statusWord(isConnected, "connected", "disconnected"), isConnected), Preview: previewLines(fmt.Sprintf("This will %s the %s server.", strings.ToLower(toggleVerb(isConnected, "disconnect", "connect")), name))},
		)
	}

	return panel
}

func panelMCPRows(names []string, connections map[string]bool) []types.RenderTableRow {
	rows := make([]types.RenderTableRow, 0, len(names))
	for _, name := range names {
		rows = append(rows, tableRow(name, statusWord(connections[name], "connected", "disconnected"), "-", boolState(connections[name], "yes", "no")))
	}
	return rows
}
