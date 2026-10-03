package commands

import (
	"fmt"

	"github.com/alliecatowo/alliecode/internal/types"
)

func buildBranchInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	ensureBranchState(state)
	branches := append([]string(nil), state.Branches...)
	return InteractivePanel{
		Command:       "branch",
		Title:         "branch panel: /branch",
		Subtitle:      fmt.Sprintf("active=%s branches=%d", defaultDash(state.ActiveBranch), len(branches)),
		HeaderIntents: branchStatusIntents(state.ActiveBranch, len(branches), state.BranchCount, state.BranchSwitchCount),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Branch status", Detail: "Inspect active branch counters", Status: "status", ApplyInput: "/branch status", ApplyMode: PanelApplySubmit, PreviewIntents: branchStatusIntents(state.ActiveBranch, len(branches), state.BranchCount, state.BranchSwitchCount)},
			{Key: "list", Section: "Overview", Label: "List branches", Detail: "Show tracked branch names", Status: "list", ApplyInput: "/branch list", ApplyMode: PanelApplySubmit, PreviewIntents: branchListIntents(state.ActiveBranch, branches)},
			{Key: "create", Section: "Actions", Label: "Create branch", Detail: "Create and switch to branch-<n>", Status: "create", ApplyInput: "/branch create", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{actionHintsIntent("Actions", hint("Create", "/branch create"))}},
		},
	}
}

func buildDiffInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	return InteractivePanel{
		Command:       "diff",
		Title:         "diff panel: /diff",
		Subtitle:      fmt.Sprintf("mode=%s entries=%d", defaultDash(state.DiffMode), len(state.DiffEntries)),
		HeaderIntents: diffStatusIntents(state.DiffMode, len(state.DiffEntries), state.DiffCount, state.DiffClearCount),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Diff status", Detail: "Show mode and counters", Status: "status", ApplyInput: "/diff status", ApplyMode: PanelApplySubmit, PreviewIntents: diffStatusIntents(state.DiffMode, len(state.DiffEntries), state.DiffCount, state.DiffClearCount)},
			{Key: "list", Section: "Overview", Label: "List diff rows", Detail: "Show deterministic diff table", Status: "list", ApplyInput: "/diff list", ApplyMode: PanelApplySubmit, PreviewIntents: []types.RenderIntent{tableIntent("Diff rows", "Tracked deterministic diff rows.", []string{"Path", "Added", "Removed", "Modified"})}},
			{Key: "mode-working", Section: "Mode", Label: "Mode working", Detail: "Switch to working tree scope", Status: statusWord(state.DiffMode == "working", "current", "mode"), ApplyInput: "/diff mode working", ApplyMode: PanelApplySubmit, PreviewIntents: diffModeIntents("working")},
			{Key: "mode-staged", Section: "Mode", Label: "Mode staged", Detail: "Switch to staged scope", Status: statusWord(state.DiffMode == "staged", "current", "mode"), ApplyInput: "/diff mode staged", ApplyMode: PanelApplySubmit, PreviewIntents: diffModeIntents("staged")},
			{Key: "mode-all", Section: "Mode", Label: "Mode all", Detail: "Switch to combined scope", Status: statusWord(state.DiffMode == "all", "current", "mode"), ApplyInput: "/diff mode all", ApplyMode: PanelApplySubmit, PreviewIntents: diffModeIntents("all")},
		},
	}
}

func buildFilesInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	files := effectiveContextFiles(state)
	return InteractivePanel{
		Command:       "files",
		Title:         "files panel: /files",
		Subtitle:      fmt.Sprintf("entries=%d adds=%d removes=%d", len(files), state.FilesAdds, state.FilesRemoves),
		HeaderIntents: filesStatusIntents(len(files), len(state.ProjectPaths), state.FilesAdds, state.FilesRemoves, state.FilesClears, state.LastFileAction, state.LastFilePath),
		Items: []InteractivePanelItem{
			{Key: "list", Section: "Overview", Label: "List files", Detail: "Show effective context-file list", Status: "list", ApplyInput: "/files list", ApplyMode: PanelApplySubmit, PreviewIntents: filesListIntents(files)},
			{Key: "status", Section: "Overview", Label: "Files status", Detail: "Show counters and last action", Status: "status", ApplyInput: "/files status", ApplyMode: PanelApplySubmit, PreviewIntents: filesStatusIntents(len(files), len(state.ProjectPaths), state.FilesAdds, state.FilesRemoves, state.FilesClears, state.LastFileAction, state.LastFilePath)},
			{Key: "clear", Section: "Actions", Label: "Clear files", Detail: "Clear all context-file rows", Status: "clear", ApplyInput: "/files clear", ApplyMode: PanelApplySubmit, PreviewIntents: filesMutationIntents("Files cleared", "", true, 0)},
		},
	}
}

func buildMemoryInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	return InteractivePanel{
		Command:       "memory",
		Title:         "memory panel: /memory",
		Subtitle:      fmt.Sprintf("entries=%d writes=%d", len(state.MemoryEntries), state.MemoryWrites),
		HeaderIntents: memoryStatusIntents(state),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Memory status", Detail: "Show memory counters", Status: "status", ApplyInput: "/memory status", ApplyMode: PanelApplySubmit, PreviewIntents: memoryStatusIntents(state)},
			{Key: "list", Section: "Overview", Label: "List memory", Detail: "Show stored deterministic memory rows", Status: "list", ApplyInput: "/memory list", ApplyMode: PanelApplySubmit, PreviewIntents: memoryListIntents(state.MemoryEntries)},
			{Key: "clear", Section: "Actions", Label: "Clear memory", Detail: "Remove all memory rows", Status: "clear", ApplyInput: "/memory clear", ApplyMode: PanelApplySubmit, PreviewIntents: memoryMutationIntents("Memory cleared", "", 0, state.MemoryClears+1)},
		},
	}
}

func buildThemeInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	theme := state.OutputStyle
	if theme == "" {
		theme = "system"
	}
	return InteractivePanel{
		Command:       "theme",
		Title:         "theme panel: /theme",
		Subtitle:      fmt.Sprintf("theme=%s set_count=%d", theme, state.ThemeSetCount),
		HeaderIntents: themeStatusIntents(theme),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Theme status", Detail: "Inspect current theme", Status: "status", ApplyInput: "/theme status", ApplyMode: PanelApplySubmit, PreviewIntents: themeStatusIntents(theme)},
			{Key: "list", Section: "Overview", Label: "List themes", Detail: "Show supported themes", Status: "list", ApplyInput: "/theme list", ApplyMode: PanelApplySubmit, PreviewIntents: themeListIntents()},
			{Key: "dark", Section: "Apply", Label: "Set dark", Detail: "Apply dark theme", Status: statusWord(theme == "dark", "current", "theme"), ApplyInput: "/theme set dark", ApplyMode: PanelApplySubmit, PreviewIntents: themeMutationIntents("Theme set", "dark", state.ThemeSetCount+1)},
			{Key: "light", Section: "Apply", Label: "Set light", Detail: "Apply light theme", Status: statusWord(theme == "light", "current", "theme"), ApplyInput: "/theme set light", ApplyMode: PanelApplySubmit, PreviewIntents: themeMutationIntents("Theme set", "light", state.ThemeSetCount+1)},
			{Key: "system", Section: "Apply", Label: "Set system", Detail: "Use system theme", Status: statusWord(theme == "system", "current", "theme"), ApplyInput: "/theme set system", ApplyMode: PanelApplySubmit, PreviewIntents: themeMutationIntents("Theme set", "system", state.ThemeSetCount+1)},
		},
	}
}

func buildOutputStyleInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	style := state.OutputFormat
	if style == "" {
		style = "default"
	}
	return InteractivePanel{
		Command:       "output-style",
		Title:         "output-style panel: /output-style",
		Subtitle:      fmt.Sprintf("style=%s set_count=%d", style, state.OutputStyleSetCount),
		HeaderIntents: outputStyleStatusIntents(style, state.OutputStyleSetCount),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Output style status", Detail: "Inspect current output style", Status: "status", ApplyInput: "/output-style status", ApplyMode: PanelApplySubmit, PreviewIntents: outputStyleStatusIntents(style, state.OutputStyleSetCount)},
			{Key: "list", Section: "Overview", Label: "List output styles", Detail: "Show supported output styles", Status: "list", ApplyInput: "/output-style list", ApplyMode: PanelApplySubmit, PreviewIntents: outputStyleListIntents()},
			{Key: "default", Section: "Apply", Label: "Set default", Detail: "Balanced output detail", Status: statusWord(style == "default", "current", "style"), ApplyInput: "/output-style set default", ApplyMode: PanelApplySubmit, PreviewIntents: outputStyleStatusIntents("default", state.OutputStyleSetCount+1)},
			{Key: "concise", Section: "Apply", Label: "Set concise", Detail: "Short output detail", Status: statusWord(style == "concise", "current", "style"), ApplyInput: "/output-style set concise", ApplyMode: PanelApplySubmit, PreviewIntents: outputStyleStatusIntents("concise", state.OutputStyleSetCount+1)},
			{Key: "explanatory", Section: "Apply", Label: "Set explanatory", Detail: "Longer output detail", Status: statusWord(style == "explanatory", "current", "style"), ApplyInput: "/output-style set explanatory", ApplyMode: PanelApplySubmit, PreviewIntents: outputStyleStatusIntents("explanatory", state.OutputStyleSetCount+1)},
			{Key: "json", Section: "Apply", Label: "Set json", Detail: "Machine-friendly output", Status: statusWord(style == "json", "current", "style"), ApplyInput: "/output-style set json", ApplyMode: PanelApplySubmit, PreviewIntents: outputStyleStatusIntents("json", state.OutputStyleSetCount+1)},
		},
	}
}

func buildPrivacyInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	return InteractivePanel{
		Command:       "privacy-settings",
		Title:         "privacy panel: /privacy-settings",
		Subtitle:      fmt.Sprintf("telemetry=%t training=%t updates=%d", state.PrivacyTelemetry, state.PrivacyTraining, state.PrivacyUpdates),
		HeaderIntents: privacyStatusIntents(state.PrivacyTelemetry, state.PrivacyTraining, state.PrivacyUpdates),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Privacy status", Detail: "Inspect telemetry and training flags", Status: "status", ApplyInput: "/privacy-settings status", ApplyMode: PanelApplySubmit, PreviewIntents: privacyStatusIntents(state.PrivacyTelemetry, state.PrivacyTraining, state.PrivacyUpdates)},
			{Key: "fields", Section: "Overview", Label: "List fields", Detail: "Show configurable privacy fields", Status: "list", ApplyInput: "/privacy-settings list", ApplyMode: PanelApplySubmit, PreviewIntents: privacyFieldsIntents()},
			{Key: "telemetry-on", Section: "Apply", Label: "Enable telemetry", Detail: "Set telemetry on", Status: statusWord(state.PrivacyTelemetry, "current", "toggle"), ApplyInput: "/privacy-settings set telemetry on", ApplyMode: PanelApplySubmit, PreviewIntents: privacySetIntents("telemetry", true, !state.PrivacyTelemetry, state.PrivacyUpdates+1)},
			{Key: "telemetry-off", Section: "Apply", Label: "Disable telemetry", Detail: "Set telemetry off", Status: statusWord(!state.PrivacyTelemetry, "current", "toggle"), ApplyInput: "/privacy-settings set telemetry off", ApplyMode: PanelApplySubmit, PreviewIntents: privacySetIntents("telemetry", false, state.PrivacyTelemetry, state.PrivacyUpdates+1)},
			{Key: "training-on", Section: "Apply", Label: "Enable training", Detail: "Set training on", Status: statusWord(state.PrivacyTraining, "current", "toggle"), ApplyInput: "/privacy-settings set training on", ApplyMode: PanelApplySubmit, PreviewIntents: privacySetIntents("training", true, !state.PrivacyTraining, state.PrivacyUpdates+1)},
			{Key: "training-off", Section: "Apply", Label: "Disable training", Detail: "Set training off", Status: statusWord(!state.PrivacyTraining, "current", "toggle"), ApplyInput: "/privacy-settings set training off", ApplyMode: PanelApplySubmit, PreviewIntents: privacySetIntents("training", false, state.PrivacyTraining, state.PrivacyUpdates+1)},
		},
	}
}

func buildUpgradeInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	return InteractivePanel{
		Command:       "upgrade",
		Title:         "upgrade panel: /upgrade",
		Subtitle:      fmt.Sprintf("requested=%t count=%d", state.UpgradeRequested, state.UpgradeCount),
		HeaderIntents: upgradeStatusIntents(state.UpgradeRequested, state.UpgradeCount, state.LastUpgradePlan),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Upgrade status", Detail: "Inspect plan guidance request state", Status: "status", ApplyInput: "/upgrade status", ApplyMode: PanelApplySubmit, PreviewIntents: upgradeStatusIntents(state.UpgradeRequested, state.UpgradeCount, state.LastUpgradePlan)},
			{Key: "max", Section: "Plans", Label: "Request max plan", Detail: "Queue max upgrade guidance", Status: statusWord(state.LastUpgradePlan == "max", "current", "plan"), ApplyInput: "/upgrade max", ApplyMode: PanelApplySubmit, PreviewIntents: upgradeRequestIntents("max", state.UpgradeCount+1)},
			{Key: "pro", Section: "Plans", Label: "Request pro plan", Detail: "Queue pro upgrade guidance", Status: statusWord(state.LastUpgradePlan == "pro", "current", "plan"), ApplyInput: "/upgrade pro", ApplyMode: PanelApplySubmit, PreviewIntents: upgradeRequestIntents("pro", state.UpgradeCount+1)},
			{Key: "team", Section: "Plans", Label: "Request team plan", Detail: "Queue team upgrade guidance", Status: statusWord(state.LastUpgradePlan == "team", "current", "plan"), ApplyInput: "/upgrade team", ApplyMode: PanelApplySubmit, PreviewIntents: upgradeRequestIntents("team", state.UpgradeCount+1)},
			{Key: "enterprise", Section: "Plans", Label: "Request enterprise plan", Detail: "Queue enterprise upgrade guidance", Status: statusWord(state.LastUpgradePlan == "enterprise", "current", "plan"), ApplyInput: "/upgrade enterprise", ApplyMode: PanelApplySubmit, PreviewIntents: upgradeRequestIntents("enterprise", state.UpgradeCount+1)},
		},
	}
}

func buildResumeInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	return InteractivePanel{
		Command:       "resume",
		Title:         "resume panel: /resume",
		Subtitle:      fmt.Sprintf("requested=%t count=%d target=%s", state.ResumeRequested, state.ResumeCount, defaultDash(state.LastResumeTarget)),
		HeaderIntents: resumeStatusIntents(state.ResumeRequested, state.ResumeCount, state.LastResumeTarget),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Resume status", Detail: "Inspect resume request counters", Status: "status", ApplyInput: "/resume status", ApplyMode: PanelApplySubmit, PreviewIntents: resumeStatusIntents(state.ResumeRequested, state.ResumeCount, state.LastResumeTarget)},
			{Key: "latest", Section: "Actions", Label: "Resume latest", Detail: "Queue resume for latest session", Status: "resume", ApplyInput: "/resume latest", ApplyMode: PanelApplySubmit, PreviewIntents: resumeRequestIntents("latest", state.ResumeCount+1)},
		},
	}
}

func buildPlanInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	return InteractivePanel{
		Command:       "plan",
		Title:         "plan panel: /plan",
		Subtitle:      fmt.Sprintf("enabled=%t opens=%d", state.PlanModeEnabled, state.PlanOpenCount),
		HeaderIntents: planStatusIntents(state.PlanModeEnabled, state.PlanEnableCount, state.PlanOpenCount, state.LastPlanDescription),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Plan status", Detail: "Inspect plan mode counters", Status: "status", ApplyInput: "/plan status", ApplyMode: PanelApplySubmit, PreviewIntents: planStatusIntents(state.PlanModeEnabled, state.PlanEnableCount, state.PlanOpenCount, state.LastPlanDescription)},
			{Key: "open", Section: "Actions", Label: "Open plan file", Detail: "Open .claude/plan.md", Status: "open", ApplyInput: "/plan open", ApplyMode: PanelApplySubmit, PreviewIntents: planMutationIntents("Plan opened", state.LastPlanDescription, false, state.PlanOpenCount+1)},
			{Key: "current", Section: "Actions", Label: "Show current plan", Detail: "Inspect current plan description", Status: "current", ApplyInput: "/plan", ApplyMode: PanelApplySubmit, PreviewIntents: planMutationIntents("Plan current", state.LastPlanDescription, false, state.PlanOpenCount)},
		},
	}
}
