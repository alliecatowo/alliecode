package commands

import (
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

func sessionInfoIntents(state *RuntimeState, remoteMode, qrAvailable bool, health sessionHealth) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Session info", "Current session transport and connectivity state.",
			field("Remote mode", boolState(remoteMode, "yes", "no")),
			field("Session ID", defaultDash(state.SessionID)),
			field("Hosted", boolState(state.SessionHosted, "yes", "no")),
			field("Connected", boolState(state.SessionConnected, "yes", "no")),
			field("Host addr", defaultDash(state.SessionHostAddr)),
			field("Connected addr", defaultDash(state.SessionConnectedAddr)),
			field("Token source", defaultDash(state.SessionTokenSource)),
			field("QR available", boolState(qrAvailable, "yes", "no")),
			field("Manager state", defaultDash(health.managerState)),
		),
		actionHintsIntent("Actions", hint("Diagnostics", "/session diagnostics")),
	}
}

func sessionURLIntents(url string) []types.RenderIntent {
	return []types.RenderIntent{
		detailRowsIntent("Session URL", "Configured remote session URL.",
			detailRow("URL", defaultDash(url), "info", "URL used for QR or remote resume flows."),
		),
	}
}

func sessionHostIntents(addr, tokenPrefix string, hostCount int) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Session host", "Started a local host transport and generated a session token.",
			field("Hosting", "yes"),
			field("Address", defaultDash(addr)),
			field("Token prefix", defaultDash(tokenPrefix)),
			field("Host count", fmt.Sprintf("%d", hostCount)),
		),
		actionListIntent("Session host actions", "Next steps after hosting a session.",
			action("Show status", "/session status", "Inspect manager and transport state.", "recommended"),
			action("Show token", "/session token show", "Inspect the redacted token prefix.", "token"),
		),
	}
}

func sessionConnectIntents(addr, tokenPrefix string, connectCount int) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Session connected", "Connected to a remote session host.",
			field("Connected", "yes"),
			field("Address", defaultDash(addr)),
			field("Token prefix", defaultDash(tokenPrefix)),
			field("Connect count", fmt.Sprintf("%d", connectCount)),
		),
		actionListIntent("Session connection actions", "Suggested follow-up actions after connecting.",
			action("Show status", "/session status", "Inspect transport and reconnect health.", "recommended"),
			action("Disconnect", "/session disconnect", "Close the current session runtime.", "disconnect"),
		),
	}
}

func sessionDisconnectIntents(wasConnected bool, previousAddr string, disconnectCount int) []types.RenderIntent {
	return []types.RenderIntent{
		detailRowsIntent("Session disconnected", "Closed active session runtime resources.",
			detailRow("Was connected", boolState(wasConnected, "yes", "no"), statusWord(wasConnected, "active", "idle"), "Whether a live session existed before disconnect."),
			detailRow("Previous address", defaultDash(previousAddr), "info", "Last connected remote address."),
			detailRow("Disconnect count", fmt.Sprintf("%d", disconnectCount), "info", "How many disconnect actions have run."),
		),
		actionListIntent("Session reconnect actions", "Useful follow-up commands after disconnect.",
			action("Host a new session", "/session host", "Start a fresh local host runtime.", "host"),
			action("Connect to a host", "/session connect <addr> <token>", "Reconnect as a client.", "connect"),
		),
	}
}

func sessionTokenIntents(source, prefix string, set bool) []types.RenderIntent {
	return []types.RenderIntent{
		detailRowsIntent("Session token", "Redacted token state for the current session.",
			detailRow("Set", boolState(set, "yes", "no"), statusWord(set, "set", "unset"), "Whether a session token is currently present."),
			detailRow("Source", defaultDash(source), "info", "Origin of the current token."),
			detailRow("Prefix", defaultDash(prefix), "info", "Redacted prefix surfaced to the UI."),
		),
		actionListIntent("Session token actions", "Follow-up commands for token management.",
			action("Show session status", "/session status", "Inspect host/connect state with the current token.", "browse"),
			action("Clear token", "/session token clear", "Remove any manually seeded token.", statusWord(set, "available", "noop")),
		),
	}
}

func sessionDiagnosticsIntents(state *RuntimeState, health sessionHealth, quickFix string) []types.RenderIntent {
	return []types.RenderIntent{diagnosticsIntent("Session diagnostics", defaultDash(state.SessionMode), "Connectivity and token health for remote sessions.", []types.RenderDiagnostic{
		diagnostic("Hosting", boolState(state.SessionHosted, "ok", "info"), boolState(state.SessionHosted, defaultDash(state.SessionHostAddr), "not hosting")),
		diagnostic("Connected", boolState(state.SessionConnected, "ok", "warn"), boolState(state.SessionConnected, defaultDash(state.SessionConnectedAddr), "not connected")),
		diagnostic("Token", boolState(stringsTrim(state.SessionToken) != "", "ok", "warn"), defaultDash(state.SessionTokenSource)),
		diagnostic("Manager", defaultDash(health.managerState), defaultDash(health.transportState)),
	}, []types.RenderActionHint{hint("Quick fix", quickFix)})}
}

func sessionRepairIntents(repairAction string, changed bool, quickFix string, repairs int) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Session repair", "Applied deterministic session repair logic.", field("Action", repairAction), field("Changed", boolState(changed, "yes", "no")), field("Repairs", fmt.Sprintf("%d", repairs))),
		actionListIntent("Session repair actions", "Suggested follow-up commands after repair.", action("Follow-up", quickFix, "Inspect or continue from the repaired session state.", statusWord(changed, "recommended", "inspect"))),
	}
}

func historyLatestIntents(entry HistoryEntry, found bool) []types.RenderIntent {
	if !found {
		return []types.RenderIntent{summaryCardIntent("Latest history", "No history entries found.", field("Found", "no"))}
	}
	return []types.RenderIntent{summaryCardIntent("Latest history", "Most recent local session history entry.", field("ID", entry.ID), field("Model", entry.Model), field("Turns", fmt.Sprintf("%d", entry.Turns)), field("Title", entry.Title))}
}

func historyListIntents(filterType, filterValue string, limit int, entries []HistoryEntry) []types.RenderIntent {
	rows := make([]types.RenderTableRow, 0, len(entries))
	for _, entry := range entries {
		rows = append(rows, tableRow(entry.ID, entry.Model, fmt.Sprintf("%d", entry.Turns), entry.Title))
	}
	return []types.RenderIntent{
		tableIntent("History list", fmt.Sprintf("filter=%s %s limit=%d", filterType, filterValue, limit), []string{"ID", "Model", "Turns", "Title"}, rows...),
	}
}

func historyShowIntents(entry HistoryEntry) []types.RenderIntent {
	return []types.RenderIntent{groupedListIntent("History entry", entry.ID, group("Meta", "Stored entry metadata.", "path: "+entry.Path, "created: "+entry.CreatedAt, "model: "+entry.Model, fmt.Sprintf("turns: %d", entry.Turns)), group("Content", "Title and summary.", "title: "+entry.Title, "summary: "+entry.Summary))}
}

func historyStatusIntents(state *RuntimeState, entryCount, modelCount int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("History status", "Current local history inventory.", field("Entries", fmt.Sprintf("%d", entryCount)), field("Views", fmt.Sprintf("%d", state.HistoryViews)), field("Last filter", fmt.Sprintf("%s %s", defaultDash(state.HistoryLastFilterType), defaultDash(state.HistoryLastFilterValue))), field("Last limit", fmt.Sprintf("%d", state.HistoryLastLimit)), field("Models", fmt.Sprintf("%d", modelCount)))}
}

func historyDoctorIntents(entryCount int, quickFix string) []types.RenderIntent {
	return []types.RenderIntent{diagnosticsIntent("History doctor", "history", "Health check for local history visibility.", []types.RenderDiagnostic{diagnostic("Entries", "info", fmt.Sprintf("%d", entryCount))}, []types.RenderActionHint{hint("Quick fix", quickFix)})}
}

func stringsTrim(v string) string {
	return strings.TrimSpace(v)
}
