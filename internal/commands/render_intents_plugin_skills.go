package commands

import (
	"fmt"

	"github.com/alliecatowo/alliecode/internal/types"
)

func pluginListIntents(state *RuntimeState) []types.RenderIntent {
	rows := make([]types.RenderTableRow, 0, len(state.PluginsInstalled))
	for _, name := range state.PluginsInstalled {
		rows = append(rows, tableRow(name, boolState(containsString(state.PluginsEnabled, name), "enabled", "disabled")))
	}
	return []types.RenderIntent{tableIntent("Plugins", "Installed plugin inventory.", []string{"Plugin", "State"}, rows...)}
}

func pluginStatusIntents(state *RuntimeState, diagnosticGroups, conflicts int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Plugin status", "Plugin inventory, diagnostics, and reload state.", field("Installed", fmt.Sprintf("%d", len(state.PluginsInstalled))), field("Enabled", fmt.Sprintf("%d", len(state.PluginsEnabled))), field("Pending reload", boolState(state.PluginReloadPending, "yes", "no")), field("Mutations", fmt.Sprintf("%d", state.PluginMutations)), field("Diagnostic groups", fmt.Sprintf("%d", diagnosticGroups)), field("Conflicts", fmt.Sprintf("%d", conflicts)))}
}

func pluginDiagnosticsIntents(state *RuntimeState, diagnosticGroups, conflicts int, quickFix string) []types.RenderIntent {
	return []types.RenderIntent{diagnosticsIntent("Plugin diagnostics", boolState(state.PluginReloadPending, "reload pending", "stable"), "Plugin install, enablement, and reload health.", []types.RenderDiagnostic{
		diagnostic("Installed", "info", fmt.Sprintf("%d", len(state.PluginsInstalled))),
		diagnostic("Enabled", "info", fmt.Sprintf("%d", len(state.PluginsEnabled))),
		diagnostic("Marketplaces", "info", fmt.Sprintf("%d", len(state.PluginMarketplaces))),
		diagnostic("Diagnostic groups", "info", fmt.Sprintf("%d", diagnosticGroups)),
		diagnostic("Conflicts", boolState(conflicts == 0, "ok", "warn"), fmt.Sprintf("%d", conflicts)),
	}, []types.RenderActionHint{hint("Quick fix", quickFix)})}
}

func pluginRepairIntents(mode string, changed, pendingReload bool, quickFix string, mutations int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Plugin repair", "Applied deterministic plugin repair flow.", field("Mode", mode), field("Changed", boolState(changed, "yes", "no")), field("Pending reload", boolState(pendingReload, "yes", "no")), field("Mutations", fmt.Sprintf("%d", mutations))), actionHintsIntent("Actions", hint("Follow-up", quickFix))}
}

func pluginMutationIntents(title, name string, mutations int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent(title, "Plugin state updated.", field("Plugin", name), field("Mutations", fmt.Sprintf("%d", mutations))), actionHintsIntent("Actions", hint("Reload plugins", "/reload-plugins"))}
}

func skillsListIntents(skills []string, views int, sources, origins map[string]string, enabled map[string]bool, conflicts int) []types.RenderIntent {
	rows := make([]types.RenderTableRow, 0, len(skills))
	enabledCount := 0
	disabledCount := 0
	for _, skill := range skills {
		state := boolState(enabled[skill], "enabled", "disabled")
		if enabled[skill] {
			enabledCount++
		} else {
			disabledCount++
		}
		rows = append(rows, tableRow(skill, state, defaultDash(sources[skill]), defaultDash(origins[skill])))
	}
	return []types.RenderIntent{
		summaryCardIntent("Skills", "Skill inventory with source and state diagnostics.",
			field("Count", fmt.Sprintf("%d", len(skills))),
			field("Enabled", fmt.Sprintf("%d", enabledCount)),
			field("Disabled", fmt.Sprintf("%d", disabledCount)),
			field("Conflicts", fmt.Sprintf("%d", conflicts)),
			field("Views", fmt.Sprintf("%d", views)),
		),
		tableIntent("Skill rows", "Current resolved skills and where they came from.", []string{"Skill", "State", "Source", "Origin"}, rows...),
	}
}

func skillsDoctorIntents(count, enabledCount, disabledCount, conflicts, views, syncCount, doctorCount int, source, quickFix string, sources, origins map[string]string) []types.RenderIntent {
	diagnostics := []types.RenderDiagnostic{
		diagnostic("Skills", "info", fmt.Sprintf("%d", count)),
		diagnostic("Enabled", "info", fmt.Sprintf("%d", enabledCount)),
		diagnostic("Disabled", boolState(disabledCount == 0, "ok", "warn"), fmt.Sprintf("%d", disabledCount)),
		diagnostic("Conflicts", boolState(conflicts == 0, "ok", "warn"), fmt.Sprintf("%d", conflicts)),
		diagnostic("Views", "info", fmt.Sprintf("%d", views)),
		diagnostic("Sync count", "info", fmt.Sprintf("%d", syncCount)),
		diagnostic("Doctor count", "info", fmt.Sprintf("%d", doctorCount)),
		diagnostic("Last source", "info", defaultDash(source)),
	}
	if len(sources) > 0 {
		diagnostics = append(diagnostics, diagnostic("Source rows", "info", fmt.Sprintf("%d", len(sources))))
	}
	if len(origins) > 0 {
		diagnostics = append(diagnostics, diagnostic("Origin rows", "info", fmt.Sprintf("%d", len(origins))))
	}
	return []types.RenderIntent{diagnosticsIntent("Skills doctor", "diagnostics", "Skill inventory and sync health.", diagnostics, []types.RenderActionHint{hint("Quick fix", quickFix)})}
}

func skillsRepairIntents(mode string, changed bool, quickFix string, count int) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Skills repair", "Applied deterministic skill repair flow.", field("Mode", mode), field("Changed", boolState(changed, "yes", "no")), field("Skills", fmt.Sprintf("%d", count))), actionHintsIntent("Actions", hint("Follow-up", quickFix))}
}

func skillsMutationIntents(title, name string, count int, changed bool) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent(title, "Skill inventory updated.", field("Skill", name), field("Changed", boolState(changed, "yes", "no")), field("Count", fmt.Sprintf("%d", count)))}
}
