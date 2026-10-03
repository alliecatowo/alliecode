package commands

import (
	"fmt"

	"github.com/alliecatowo/alliecode/internal/types"
)

func buildStatusInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	syncProviderReadiness(state)
	runtime := statusRuntimeSnapshot(state)
	selection := RuntimeSelectionTruth(state)
	return InteractivePanel{
		Command:       "status",
		Title:         "status panel: /status",
		Subtitle:      fmt.Sprintf("provider=%s model=%s turns=%d", normalizeToken(selection.ProviderName), normalizeToken(selection.ModelName), runtime.Turns),
		HeaderIntents: statusReportIntents(state, runtime),
		Items: []InteractivePanelItem{
			{Key: "overview", Section: "Overview", Label: "Runtime overview", Detail: "Show the deterministic session/runtime snapshot", Status: statusWord(selection.ProviderReady, "ready", "warn"), ApplyInput: "/status", ApplyMode: PanelApplySubmit, PreviewIntents: statusReportIntents(state, runtime), Preview: previewLines(fmt.Sprintf("Model: %s", normalizeToken(selection.ModelName)), fmt.Sprintf("Provider ready: %t", selection.ProviderReady), fmt.Sprintf("Turns: %d", runtime.Turns), fmt.Sprintf("Tasks running: %d", runtime.TasksRunning))},
			{Key: "diagnostics", Section: "Diagnostics", Label: "Status diagnostics", Detail: "Inspect corrective loops and runtime checks", Status: "diagnostics", ApplyInput: "/status diagnostics", ApplyMode: PanelApplySubmit, PreviewIntents: statusDiagnosticsIntents(state, correctiveLoopsForState(state)), Preview: previewLines("/status diagnostics emits corrective loops for the current runtime state.")},
			{Key: "doctor", Section: "Diagnostics", Label: "Open doctor", Detail: "Pivot to the broader diagnostic report", Status: "related", ApplyInput: "/doctor", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{actionListIntent("Related diagnostics", "Commands related to runtime status.", action("Open doctor", "/doctor", "Pivot to the broader diagnostic report.", "related"))}, Preview: previewLines("Use /doctor for the broader cross-area diagnostic report.")},
		},
	}
}
