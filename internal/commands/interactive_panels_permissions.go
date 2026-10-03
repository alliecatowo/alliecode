package commands

import "fmt"

func buildPermissionsInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	rules := permissionRulesFromStrings(state.PermissionRules)
	denials := append([]PermissionDenial(nil), state.PermissionDenials...)
	mode := modeString(state.PermissionMode)

	panel := InteractivePanel{
		Command:       "permissions",
		Title:         "permissions panel: /permissions",
		Subtitle:      fmt.Sprintf("mode=%s rules=%d denials=%d", mode, len(rules), len(denials)),
		HeaderIntents: permissionsSummaryIntents(state, denialGroupReasons(groupPermissionDenials(denials))),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Permissions summary", Detail: "Inspect mode, rule sources, and denial counts", Status: mode, ApplyInput: "/permissions status", ApplyMode: PanelApplySubmit, PreviewIntents: permissionsSummaryIntents(state, denialGroupReasons(groupPermissionDenials(denials))), Preview: previewLines(fmt.Sprintf("Current mode: %s", mode), fmt.Sprintf("Rules: %d", len(rules)), fmt.Sprintf("Denials: %d", len(denials)))},
			{Key: "plan", Section: "Modes", Label: "Set plan mode", Detail: "Ask for approval before applying edits", Status: statusWord(mode == "plan", "current", "mode"), ApplyInput: "/permissions plan", ApplyMode: PanelApplySubmit, PreviewIntents: permissionsModeIntents("plan"), Preview: previewLines("Plan mode keeps changes gated until approval.")},
			{Key: "default", Section: "Modes", Label: "Set default mode", Detail: "Use the repo-default permission behavior", Status: statusWord(mode == "default", "current", "mode"), ApplyInput: "/permissions default", ApplyMode: PanelApplySubmit, PreviewIntents: permissionsModeIntents("default"), Preview: previewLines("Default mode follows configured policy and prompt rules.")},
			{Key: "auto", Section: "Modes", Label: "Set auto mode", Detail: "Automatically accept edits without per-step prompts", Status: statusWord(mode == "auto", "current", "mode"), ApplyInput: "/permissions auto", ApplyMode: PanelApplySubmit, PreviewIntents: permissionsModeIntents("auto"), Preview: previewLines("Auto mode removes most edit confirmations.")},
			{Key: "bypass", Section: "Modes", Label: "Set bypass mode", Detail: "Bypass permission gating entirely", Status: statusWord(mode == "bypass", "current", "danger"), ApplyInput: "/permissions bypass", ApplyMode: PanelApplySubmit, PreviewIntents: permissionsModeIntents("bypass"), Preview: previewLines("Bypass mode is the least restrictive option.")},
			{Key: "rules", Section: "Policies", Label: "View rules", Detail: fmt.Sprintf("Review %d effective permission rules", len(rules)), Status: statusWord(len(rules) > 0, "rules", "empty"), ApplyInput: "/permissions rules", ApplyMode: PanelApplySubmit, PreviewIntents: permissionRulesIntents(state.PermissionRules), Preview: previewLines(fmt.Sprintf("Rule count: %d", len(rules)), "Rules are grouped by policy, user, project, and session sources.")},
			{Key: "denials", Section: "Policies", Label: "View denials", Detail: fmt.Sprintf("Inspect %d denied tool requests", len(denials)), Status: statusWord(len(denials) > 0, "attention", "clear"), ApplyInput: "/permissions denials", ApplyMode: PanelApplySubmit, PreviewIntents: permissionDenialsIntents(denials, denialGroupReasons(groupPermissionDenials(denials))), Preview: previewLines(fmt.Sprintf("Denied requests: %d", len(denials)), "Use this to inspect repeated failures before retrying.")},
			{Key: "retry-denials", Section: "Policies", Label: "Retry denied commands", Detail: "Show deterministic retry commands for denied requests", Status: statusWord(len(permissionRetryCommands(denials)) > 0, "retry", "empty"), ApplyInput: "/permissions retry-denials", ApplyMode: PanelApplySubmit, PreviewIntents: permissionRetryIntents(permissionRetryCommands(denials)), Preview: previewLines("This lists retry-ready commands for the current denial history.")},
			{Key: "doctor", Section: "Diagnostics", Label: "Permission loops", Detail: "Inspect corrective loops involving permissions/settings", Status: "diagnostics", ApplyInput: "/permissions doctor", ApplyMode: PanelApplySubmit, PreviewIntents: correctiveLoopIntents("Permission loops", "Permission doctor highlights loops and next actions.", correctiveLoopsForState(state)), Preview: previewLines("Permission doctor highlights loops and next actions.")},
		},
	}

	return panel
}
