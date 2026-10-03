package commands

import (
	"fmt"

	"github.com/alliecatowo/alliecode/internal/types"
)

func buildBridgeInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	return InteractivePanel{
		Command:  "bridge",
		Title:    "bridge panel: /bridge",
		Subtitle: fmt.Sprintf("enabled=%t transitions=%d", state.BridgeEnabled, state.BridgeTransitions),
		HeaderIntents: []types.RenderIntent{
			summaryCardIntent("Bridge", "Bridge routing mode.", field("Enabled", boolState(state.BridgeEnabled, "yes", "no")), field("Transitions", itoa(state.BridgeTransitions))),
		},
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Bridge status", Detail: "Inspect bridge runtime state", Status: "status", ApplyInput: "/bridge status", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{summaryCardIntent("Bridge", "Bridge routing mode.", field("Enabled", boolState(state.BridgeEnabled, "yes", "no")), field("Transitions", itoa(state.BridgeTransitions)))}},
			{Key: "on", Section: "Actions", Label: "Enable bridge", Detail: "Route through bridge mode", Status: statusWord(state.BridgeEnabled, "current", "toggle"), ApplyInput: "/bridge on", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{detailRowsIntent("Bridge update", "Enable bridge mode.", detailRow("Enabled", "yes", "enabled", "Bridge mode after update."))}},
			{Key: "off", Section: "Actions", Label: "Disable bridge", Detail: "Use direct mode", Status: statusWord(!state.BridgeEnabled, "current", "toggle"), ApplyInput: "/bridge off", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{detailRowsIntent("Bridge update", "Disable bridge mode.", detailRow("Enabled", "no", "disabled", "Bridge mode after update."))}},
			{Key: "toggle", Section: "Actions", Label: "Toggle bridge", Detail: "Flip current bridge state", Status: "toggle", ApplyInput: "/bridge toggle", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{actionHintsIntent("Bridge actions", hint("Toggle", "/bridge toggle"))}},
		},
	}
}

func buildSandboxToggleInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	mode := state.SandboxMode
	if mode == "" {
		mode = "off"
	}
	return InteractivePanel{
		Command:       "sandbox-toggle",
		Title:         "sandbox-toggle panel: /sandbox-toggle",
		Subtitle:      fmt.Sprintf("mode=%s", mode),
		HeaderIntents: []types.RenderIntent{summaryCardIntent("Sandbox toggle", "Quick sandbox mode switch.", field("Mode", mode))},
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Sandbox status", Detail: "Inspect sandbox quick-toggle mode", Status: "status", ApplyInput: "/sandbox-toggle status", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{summaryCardIntent("Sandbox toggle", "Quick sandbox mode switch.", field("Mode", mode))}},
			{Key: "on", Section: "Actions", Label: "Enable sandbox", Detail: "Set workspace-write mode", Status: statusWord(mode == "workspace-write", "current", "mode"), ApplyInput: "/sandbox-toggle on", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{detailRowsIntent("Sandbox mode", "Enable workspace-write mode.", detailRow("Mode", "workspace-write", "mode", "Sandbox mode after update."))}},
			{Key: "off", Section: "Actions", Label: "Disable sandbox", Detail: "Set mode off", Status: statusWord(mode == "off", "current", "mode"), ApplyInput: "/sandbox-toggle off", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{detailRowsIntent("Sandbox mode", "Disable sandbox mode.", detailRow("Mode", "off", "mode", "Sandbox mode after update."))}},
			{Key: "toggle", Section: "Actions", Label: "Toggle sandbox", Detail: "Flip between off/workspace-write", Status: "toggle", ApplyInput: "/sandbox-toggle toggle", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{actionHintsIntent("Sandbox actions", hint("Toggle", "/sandbox-toggle toggle"))}},
		},
	}
}

func buildMockLimitsInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	return InteractivePanel{
		Command:       "mock-limits",
		Title:         "mock-limits panel: /mock-limits",
		Subtitle:      fmt.Sprintf("enabled=%t value=%d", state.MockLimitsEnabled, state.MockLimitsValue),
		HeaderIntents: []types.RenderIntent{summaryCardIntent("Mock limits", "Deterministic quota override mode.", field("Enabled", boolState(state.MockLimitsEnabled, "yes", "no")), field("Value", itoa(state.MockLimitsValue)))},
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Mock limits status", Detail: "Inspect mock quota overrides", Status: "status", ApplyInput: "/mock-limits status", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{summaryCardIntent("Mock limits", "Deterministic quota override mode.", field("Enabled", boolState(state.MockLimitsEnabled, "yes", "no")), field("Value", itoa(state.MockLimitsValue)))}},
			{Key: "on", Section: "Actions", Label: "Enable mock limits", Detail: "Turn on deterministic limit overrides", Status: statusWord(state.MockLimitsEnabled, "current", "toggle"), ApplyInput: "/mock-limits on", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{detailRowsIntent("Mock limits", "Enable deterministic mock limits.", detailRow("Enabled", "yes", "enabled", "Mock limits after update."))}},
			{Key: "off", Section: "Actions", Label: "Disable mock limits", Detail: "Turn off overrides", Status: statusWord(!state.MockLimitsEnabled, "current", "toggle"), ApplyInput: "/mock-limits off", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{detailRowsIntent("Mock limits", "Disable deterministic mock limits.", detailRow("Enabled", "no", "disabled", "Mock limits after update."))}},
		},
	}
}

func buildOnboardingInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	return InteractivePanel{
		Command:       "onboarding",
		Title:         "onboarding panel: /onboarding",
		Subtitle:      fmt.Sprintf("completed=%t runs=%d", state.OnboardingCompleted, state.OnboardingRuns),
		HeaderIntents: []types.RenderIntent{summaryCardIntent("Onboarding", "Startup onboarding state.", field("Completed", boolState(state.OnboardingCompleted, "yes", "no")), field("Runs", itoa(state.OnboardingRuns)))},
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Onboarding status", Detail: "Inspect startup onboarding status", Status: "status", ApplyInput: "/onboarding status", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{summaryCardIntent("Onboarding", "Startup onboarding state.", field("Completed", boolState(state.OnboardingCompleted, "yes", "no")), field("Runs", itoa(state.OnboardingRuns)))}},
			{Key: "run", Section: "Actions", Label: "Run onboarding", Detail: "Mark onboarding complete", Status: statusWord(state.OnboardingCompleted, "current", "state"), ApplyInput: "/onboarding run", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{detailRowsIntent("Onboarding", "Mark onboarding completed.", detailRow("Completed", "yes", "done", "Onboarding completion flag."))}},
			{Key: "reset", Section: "Actions", Label: "Reset onboarding", Detail: "Mark onboarding incomplete", Status: statusWord(!state.OnboardingCompleted, "current", "state"), ApplyInput: "/onboarding reset", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{detailRowsIntent("Onboarding", "Mark onboarding incomplete.", detailRow("Completed", "no", "pending", "Onboarding completion flag."))}},
		},
	}
}

func buildRemoteSetupInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	return InteractivePanel{
		Command:       "remote-setup",
		Title:         "remote-setup panel: /remote-setup",
		Subtitle:      fmt.Sprintf("connected=%t count=%d", state.WebSetupConnected, state.WebSetupCount),
		HeaderIntents: []types.RenderIntent{summaryCardIntent("Remote setup", "Web setup connection state.", field("Connected", boolState(state.WebSetupConnected, "yes", "no")), field("Updates", itoa(state.WebSetupCount)))},
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Remote setup status", Detail: "Inspect web setup connection", Status: "status", ApplyInput: "/remote-setup status", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{summaryCardIntent("Remote setup", "Web setup connection state.", field("Connected", boolState(state.WebSetupConnected, "yes", "no")), field("Updates", itoa(state.WebSetupCount)))}},
			{Key: "connect", Section: "Actions", Label: "Connect remote setup", Detail: "Enable web setup link", Status: statusWord(state.WebSetupConnected, "current", "state"), ApplyInput: "/remote-setup connect", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{detailRowsIntent("Remote setup", "Connect remote setup flow.", detailRow("Connected", "yes", "connected", "Connection state after update."))}},
			{Key: "disconnect", Section: "Actions", Label: "Disconnect remote setup", Detail: "Disable web setup link", Status: statusWord(!state.WebSetupConnected, "current", "state"), ApplyInput: "/remote-setup disconnect", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{detailRowsIntent("Remote setup", "Disconnect remote setup flow.", detailRow("Connected", "no", "disconnected", "Connection state after update."))}},
		},
	}
}
