package commands

import (
	"fmt"
	"strings"
)

func buildSessionInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	health := sessionHealthSummary(state.SessionManager)
	panel := InteractivePanel{
		Command:       "session",
		Title:         "session panel: /session",
		Subtitle:      fmt.Sprintf("mode=%s connected=%t hosting=%t", normalizeToken(state.SessionMode), state.SessionConnected, state.SessionHosted),
		HeaderIntents: sessionInfoIntents(state, strings.EqualFold(strings.TrimSpace(state.TransportMode), "remote"), strings.TrimSpace(state.RemoteSessionURL) != "", health),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Session overview", Detail: "Show remote session URL, token state, and health", Status: statusWord(state.SessionConnected || state.SessionHosted, "active", "idle"), ApplyInput: "/session status", ApplyMode: PanelApplySubmit, PreviewIntents: sessionInfoIntents(state, strings.EqualFold(strings.TrimSpace(state.TransportMode), "remote"), strings.TrimSpace(state.RemoteSessionURL) != "", health), Preview: previewLines(fmt.Sprintf("URL set: %t", strings.TrimSpace(state.RemoteSessionURL) != ""), fmt.Sprintf("Token set: %t", strings.TrimSpace(state.SessionToken) != ""), fmt.Sprintf("Manager state: %s", normalizeToken(health.managerState)))},
			{Key: "diagnostics", Section: "Diagnostics", Label: "Session diagnostics", Detail: "Inspect transport, reconnect, and repair counters", Status: "diagnostics", ApplyInput: "/session diagnostics", ApplyMode: PanelApplySubmit, PreviewIntents: sessionDiagnosticsIntents(state, health, "/session host"), Preview: previewLines(fmt.Sprintf("Transport state: %s", normalizeToken(health.transportState)), fmt.Sprintf("Reconnect count: %d", health.reconnectCount))},
			{Key: "token", Section: "Access", Label: "Show session token", Detail: "Inspect token source and redacted prefix", Status: statusWord(strings.TrimSpace(state.SessionToken) != "", "set", "unset"), ApplyInput: "/session token show", ApplyMode: PanelApplySubmit, PreviewIntents: sessionTokenIntents(state.SessionTokenSource, state.SessionTokenPrefix, strings.TrimSpace(state.SessionToken) != ""), Preview: previewLines(fmt.Sprintf("Token source: %s", normalizeToken(state.SessionTokenSource)), fmt.Sprintf("Token prefix: %s", normalizeToken(state.SessionTokenPrefix)))},
			{Key: "host", Section: "Actions", Label: "Host session", Detail: "Start hosting a local session on the default address", Status: "host", ApplyInput: "/session host", ApplyMode: PanelApplySubmit, PreviewIntents: sessionHostIntents(defaultDash(state.SessionHostAddr), defaultDash(state.SessionTokenPrefix), state.SessionHostCount), Preview: previewLines("/session host starts a local host transport and seeds a session token.")},
			{Key: "repair", Section: "Actions", Label: "Repair session state", Detail: "Apply deterministic repair flow", Status: "repair", ApplyInput: "/session repair", ApplyMode: PanelApplySubmit, PreviewIntents: sessionRepairIntents("auto", false, "/session status", state.SessionRepairCount), Preview: previewLines("/session repair chooses a token/disconnect/noop fix based on current runtime state.")},
		},
	}

	if state.SessionConnected || state.SessionHosted {
		panel.Items = append(panel.Items, InteractivePanelItem{Key: "disconnect", Section: "Actions", Label: "Disconnect session", Detail: "Close active session runtime resources", Status: "disconnect", ApplyInput: "/session disconnect", ApplyMode: PanelApplySubmit, PreviewIntents: sessionDisconnectIntents(true, state.SessionConnectedAddr, state.SessionDisconnectCount), Preview: previewLines("This shuts down any current host/client transport runtime.")})
	}

	return panel
}
