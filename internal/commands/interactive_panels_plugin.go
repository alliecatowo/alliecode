package commands

import (
	"fmt"

	"github.com/alliecatowo/alliecode/internal/types"
)

func buildPluginInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	normalizePluginState(state)
	panel := InteractivePanel{
		Command:       "plugin",
		Title:         "plugin panel: /plugin",
		Subtitle:      fmt.Sprintf("installed=%d enabled=%d pending_reload=%t", len(state.PluginsInstalled), len(state.PluginsEnabled), state.PluginReloadPending),
		HeaderIntents: pluginStatusIntents(state, 0, 0),
		Items: []InteractivePanelItem{
			{Key: "list", Section: "Overview", Label: "Plugin inventory", Detail: "Show installed plugins, enabled set, diagnostics, and conflicts", Status: statusWord(len(state.PluginsInstalled) > 0, "inventory", "empty"), ApplyInput: "/plugin list", ApplyMode: PanelApplySubmit, PreviewIntents: pluginListIntents(state), Preview: previewLines(fmt.Sprintf("Installed plugins: %d", len(state.PluginsInstalled)), fmt.Sprintf("Enabled plugins: %d", len(state.PluginsEnabled)))},
			{Key: "status", Section: "Overview", Label: "Plugin status", Detail: "Summarize plugin counts and pending reload state", Status: statusWord(state.PluginReloadPending, "reload", "clean"), ApplyInput: "/plugin status", ApplyMode: PanelApplySubmit, PreviewIntents: pluginStatusIntents(state, 0, 0), Preview: previewLines(fmt.Sprintf("Pending reload: %t", state.PluginReloadPending), fmt.Sprintf("Mutations: %d", state.PluginMutations))},
			{Key: "doctor", Section: "Diagnostics", Label: "Plugin doctor", Detail: "Inspect plugin diagnostics and quick fixes", Status: "diagnostics", ApplyInput: "/plugin doctor", ApplyMode: PanelApplySubmit, PreviewIntents: pluginDiagnosticsIntents(state, 0, 0, "/plugin repair"), Preview: previewLines("Use plugin doctor when installs, enablement, or reload state looks wrong.")},
			{Key: "diagnostics", Section: "Diagnostics", Label: "Plugin diagnostics", Detail: "Render deterministic plugin diagnostic rows", Status: "diagnostics", ApplyInput: "/plugin diagnostics", ApplyMode: PanelApplySubmit, PreviewIntents: pluginDiagnosticsIntents(state, 0, 0, "/plugin repair"), Preview: previewLines("/plugin diagnostics includes conflicts, reload state, and inventories.")},
			{Key: "repair", Section: "Diagnostics", Label: "Repair plugin state", Detail: "Apply deterministic plugin repair flow", Status: "repair", ApplyInput: "/plugin repair", ApplyMode: PanelApplySubmit, PreviewIntents: pluginRepairIntents("auto", false, state.PluginReloadPending, "/plugin status", state.PluginMutations), Preview: previewLines("/plugin repair picks a deterministic repair mode from runtime state.")},
		},
	}

	if state.PluginReloadPending {
		panel.Items = append(panel.Items, InteractivePanelItem{Key: "reload", Section: "Mutations", Label: "Reload plugins", Detail: "Apply pending plugin mutations", Status: "pending", ApplyInput: "/reload-plugins", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{actionListIntent("Reload plugins", "Apply pending plugin mutations.", action("Reload plugins", "/reload-plugins", "Apply pending install/enable/disable/remove changes.", "pending"))}, Preview: previewLines("Reload applies pending install/enable/disable/remove changes.")})
	}

	for _, name := range state.PluginsInstalled {
		enabled := containsString(state.PluginsEnabled, name)
		panel.Items = append(panel.Items,
			InteractivePanelItem{Key: name + ":toggle", Section: "Installed plugins", Label: fmt.Sprintf("%s plugin %s", toggleVerb(enabled, "Disable", "Enable"), name), Detail: "Toggle whether this plugin is active", Status: statusWord(enabled, "enabled", "disabled"), ApplyInput: toggleCommand(enabled, "/plugin disable "+name, "/plugin enable "+name), ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{detailRowsIntent("Plugin toggle", "Plugin enablement summary.", detailRow("Plugin", name, statusWord(enabled, "enabled", "disabled"), "Affected plugin."), detailRow("Enabled", boolState(enabled, "yes", "no"), statusWord(enabled, "enabled", "disabled"), "Current plugin state."))}, Preview: previewLines(fmt.Sprintf("Plugin: %s", name), fmt.Sprintf("Enabled: %t", enabled))},
			InteractivePanelItem{Key: name + ":remove", Section: "Installed plugins", Label: fmt.Sprintf("Remove plugin %s", name), Detail: "Uninstall this plugin from the workspace", Status: "remove", ApplyInput: "/plugin remove " + name, ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{actionListIntent("Plugin removal", "Remove a plugin and mark reload pending.", action("Remove plugin", "/plugin remove "+name, "Uninstall this plugin from the workspace.", "remove"))}, Preview: previewLines(fmt.Sprintf("This will remove the %s plugin and mark reload pending.", name))},
		)
	}

	return panel
}
