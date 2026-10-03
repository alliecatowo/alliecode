package commands

import (
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/providers"
	"github.com/alliecatowo/alliecode/internal/types"
)

func buildModelInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	selection := RuntimeSelectionTruth(state)
	current := strings.TrimSpace(selection.ModelRef)
	if current == "" && strings.TrimSpace(selection.ModelName) != "" {
		if strings.TrimSpace(selection.ProviderName) != "" {
			current = strings.TrimSpace(selection.ProviderName) + "/" + strings.TrimSpace(selection.ModelName)
		} else {
			current = strings.TrimSpace(selection.ModelName)
		}
	}
	if current == "" {
		current = "none"
	}

	panel := InteractivePanel{
		Command:       "model",
		Title:         "model panel: /model",
		Subtitle:      fmt.Sprintf("current=%s provider=%s", current, normalizeToken(selection.ProviderName)),
		HeaderIntents: modelStatusIntents(selection.ProviderName, selection.ModelName, currentModelCapabilities(selection.ProviderName, selection.ModelName), "/model list", selection.ProviderReady),
		Items: []InteractivePanelItem{
			{
				Key:            "status",
				Section:        "Overview",
				Label:          "Current model status",
				Detail:         fmt.Sprintf("Inspect the active provider/model pair (%s)", current),
				Status:         statusWord(strings.TrimSpace(selection.ModelRef) != "", "active", "unset"),
				ApplyInput:     "/model",
				ApplyMode:      PanelApplySubmit,
				PreviewIntents: modelStatusIntents(selection.ProviderName, selection.ModelName, currentModelCapabilities(selection.ProviderName, selection.ModelName), "/model list", selection.ProviderReady),
				Preview: previewLines(
					"/model returns the current runtime selection and capability summary.",
					fmt.Sprintf("Current selection: %s", current),
				),
			},
			{
				Key:            "list-all",
				Section:        "Overview",
				Label:          "Browse all providers",
				Detail:         "Show the provider/model matrix and context windows",
				Status:         "browse",
				ApplyInput:     "/model list all",
				ApplyMode:      PanelApplySubmit,
				PreviewIntents: modelListAllIntents(),
				Preview:        previewLines("/model list all shows every supported provider/model combination."),
			},
			{
				Key:            "doctor",
				Section:        "Diagnostics",
				Label:          "Model doctor",
				Detail:         "Check whether the active provider/model pair is usable",
				Status:         statusWord(selection.ProviderReady && strings.TrimSpace(selection.ModelName) != "", "ready", "needs-fix"),
				ApplyInput:     "/model doctor",
				ApplyMode:      PanelApplySubmit,
				PreviewIntents: modelDoctorIntents(selection.ProviderName, selection.ModelName, selection.ProviderReady, "/model repair "+current),
				Preview:        previewLines("/model doctor reports readiness and suggests a repair command when unset or invalid."),
			},
		},
	}

	if provider := strings.TrimSpace(selection.ProviderName); provider != "" {
		repairTarget := provider + "/<model>"
		if strings.TrimSpace(selection.ModelName) != "" {
			repairTarget = provider + "/" + strings.TrimSpace(selection.ModelName)
		}
		panel.Items = append(panel.Items, InteractivePanelItem{
			Key:            "repair",
			Section:        "Diagnostics",
			Label:          "Repair current selection",
			Detail:         fmt.Sprintf("Re-apply the current provider/model pair (%s)", repairTarget),
			Status:         "repair",
			ApplyInput:     "/model repair " + repairTarget,
			ApplyMode:      PanelApplySubmit,
			PreviewIntents: modelRepairIntents(provider, strings.TrimSpace(selection.ModelName), selection.ProviderReady),
			Preview:        previewLines("/model repair re-applies a known selection and refreshes runtime state."),
		})
	}

	for _, providerName := range providers.SupportedProviderNames() {
		ready := ProviderReadyForState(providerName, state)
		for _, model := range providers.ListModelsByProvider(providerName) {
			ref := providerName + "/" + model.Model
			status := statusWord(strings.EqualFold(ref, strings.TrimSpace(selection.ModelRef)), "current", statusWord(ready, "ready", "auth"))
			panel.Items = append(panel.Items, InteractivePanelItem{
				Key:            ref,
				Section:        providerName,
				Label:          ref,
				Detail:         fmt.Sprintf("ctx=%d caps=%s", model.ContextWindow, model.CapabilitySummary()),
				Status:         status,
				ApplyInput:     "/model " + ref,
				ApplyMode:      PanelApplySubmit,
				PreviewIntents: []types.RenderIntent{optionListIntent("Model target", "Structured model selection preview.", option(ref, fmt.Sprintf("ctx=%d caps=%s", model.ContextWindow, model.CapabilitySummary()), status, "/model "+ref, strings.EqualFold(ref, strings.TrimSpace(selection.ModelRef))))},
				Preview: previewLines(
					fmt.Sprintf("Provider: %s", providerName),
					fmt.Sprintf("Model: %s", model.Model),
					fmt.Sprintf("Context window: %d", model.ContextWindow),
					fmt.Sprintf("Capabilities: %s", model.CapabilitySummary()),
					fmt.Sprintf("Provider ready: %t", ready),
				),
			})
		}
	}

	return panel
}

func currentModelCapabilities(providerName, modelName string) string {
	providerName = strings.TrimSpace(providerName)
	modelName = strings.TrimSpace(modelName)
	if providerName == "" || modelName == "" {
		return "-"
	}
	for _, model := range providers.ListModelsByProvider(providerName) {
		if strings.EqualFold(model.Model, modelName) {
			return model.CapabilitySummary()
		}
	}
	return "-"
}
