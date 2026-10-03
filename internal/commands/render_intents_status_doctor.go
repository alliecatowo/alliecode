package commands

import (
	"fmt"

	"github.com/alliecatowo/alliecode/internal/types"
)

func statusReportIntents(state *RuntimeState, runtime types.AgentRuntimeSnapshot) []types.RenderIntent {
	selection := RuntimeSelectionTruth(state)
	hints := []types.RenderActionHint{hint("Diagnostics", "/status diagnostics")}
	if next := RuntimeSelectionNextAction(state); next != "" && next != "/status diagnostics" {
		hints = append([]types.RenderActionHint{hint("Next", next)}, hints...)
	}
	return []types.RenderIntent{
		summaryCardIntent("Status", "Deterministic runtime snapshot for the current session.",
			field("Model", defaultDash(selection.ModelName)),
			field("Provider", defaultDash(selection.ProviderName)),
			field("Logged in", boolState(selection.LoggedIn, "yes", "no")),
			field("Provider ready", boolState(selection.ProviderReady, "yes", "no")),
			field("Permissions", modeString(state.PermissionMode)),
			field("Turns", fmt.Sprintf("%d", runtime.Turns)),
			field("Tasks", fmt.Sprintf("%d total / %d running / %d done", runtime.TasksTotal, runtime.TasksRunning, runtime.TasksCompleted)),
		),
		actionHintsIntent("Actions", hints...),
	}
}

func statusDiagnosticsIntents(state *RuntimeState, loops []correctiveLoop) []types.RenderIntent {
	selection := RuntimeSelectionTruth(state)
	diags := []types.RenderDiagnostic{
		diagnostic("Provider", "info", defaultDash(selection.ProviderName)),
		diagnostic("Model", "info", defaultDash(selection.ModelName)),
		diagnostic("Permission mode", "info", modeString(state.PermissionMode)),
		diagnostic("Loop count", "info", fmt.Sprintf("%d", len(loops))),
	}
	for _, loop := range loops {
		diags = append(diags, diagnostic(loop.Area, "warn", fmt.Sprintf("state=%s next=%s action=%s", loop.State, loop.Next, loop.Action)))
	}
	return []types.RenderIntent{diagnosticsIntent("Status diagnostics", "diagnostics", "Corrective-loop oriented runtime checks.", diags, nil)}
}

func doctorReportIntents(status string, sections []doctorSection) []types.RenderIntent {
	items := make([]types.RenderChecklistItem, 0, len(sections))
	for _, section := range sections {
		items = append(items, checklistItem(section.Name, section.Status == "ok", fmt.Sprintf("%s (%d checks)", section.Status, len(section.Checks))))
	}
	return []types.RenderIntent{checklistIntent("Doctor report", fmt.Sprintf("Overall status: %s", status), items...)}
}

func doctorFixPlanIntents(status string, warnSections int, quickFixes []string, loops []correctiveLoop) []types.RenderIntent {
	fixItems := make([]types.RenderChecklistItem, 0, len(quickFixes))
	for _, quickFix := range quickFixes {
		fixItems = append(fixItems, checklistItem(quickFix, false, "suggested remediation"))
	}
	hints := make([]types.RenderActionHint, 0, len(quickFixes))
	for _, quickFix := range quickFixes {
		hints = append(hints, hint("Run", quickFix))
	}
	return []types.RenderIntent{
		diagnosticsIntent("Doctor fix plan", status, fmt.Sprintf("%d warning sections need attention.", warnSections), nil, hints),
		checklistIntent("Suggested fixes", "Follow these actions in order.", fixItems...),
		checklistIntent("Corrective loops", "Loop hotspots surfaced by doctor.", correctiveLoopChecklistItems(loops)...),
	}
}

func correctiveLoopChecklistItems(loops []correctiveLoop) []types.RenderChecklistItem {
	items := make([]types.RenderChecklistItem, 0, len(loops))
	for _, loop := range loops {
		items = append(items, checklistItem(loop.Area, false, fmt.Sprintf("state=%s next=%s action=%s", loop.State, loop.Next, loop.Action)))
	}
	return items
}
