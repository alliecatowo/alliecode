package commands

import (
	"fmt"
	"sort"
	"strings"
)

func buildCompactInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	mode := strings.TrimSpace(state.CompactMode)
	if mode == "" {
		mode = "auto"
	}
	return InteractivePanel{
		Command:       "compact",
		Title:         "compact panel: /compact",
		Subtitle:      fmt.Sprintf("mode=%s requested=%t count=%d", mode, state.CompactRequested, state.CompactCount),
		HeaderIntents: compactIntents(mode, state.CompactRequested, state.CompactCount, state.LastCompactTarget),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Compact status", Detail: "Inspect compaction mode and queue state", Status: "status", ApplyInput: "/compact status", ApplyMode: PanelApplySubmit, PreviewIntents: compactIntents(mode, state.CompactRequested, state.CompactCount, state.LastCompactTarget)},
			{Key: "now", Section: "Actions", Label: "Compact now", Detail: "Queue compaction for next model turn", Status: statusWord(state.CompactRequested, "queued", "run"), ApplyInput: "/compact now", ApplyMode: PanelApplySubmit, PreviewIntents: compactIntents(mode, true, state.CompactCount+1, "now")},
			{Key: "auto", Section: "Mode", Label: "Set auto mode", Detail: "Enable automatic compaction behavior", Status: statusWord(mode == "auto", "current", "mode"), ApplyInput: "/compact auto", ApplyMode: PanelApplySubmit, PreviewIntents: compactIntents("auto", false, state.CompactCount, state.LastCompactTarget)},
			{Key: "off", Section: "Mode", Label: "Set off mode", Detail: "Disable automatic compaction behavior", Status: statusWord(mode == "off", "current", "mode"), ApplyInput: "/compact off", ApplyMode: PanelApplySubmit, PreviewIntents: compactIntents("off", false, state.CompactCount, state.LastCompactTarget)},
		},
	}
}

func buildTasksInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	panel := InteractivePanel{
		Command:       "tasks",
		Title:         "tasks panel: /tasks",
		Subtitle:      fmt.Sprintf("open=%d completed=%d", len(state.Tasks), state.TasksCompleted),
		HeaderIntents: tasksStatusIntents(state.Tasks, state.TasksCompleted),
		Items: []InteractivePanelItem{
			{Key: "list", Section: "Overview", Label: "List tasks", Detail: "Show open deterministic tasks", Status: "list", ApplyInput: "/tasks list", ApplyMode: PanelApplySubmit, PreviewIntents: tasksListIntents(state.Tasks, state.TasksCompleted)},
			{Key: "clear", Section: "Actions", Label: "Clear tasks", Detail: "Remove all open task rows", Status: statusWord(len(state.Tasks) == 0, "empty", "clear"), ApplyInput: "/tasks clear", ApplyMode: PanelApplySubmit, PreviewIntents: tasksMutationIntents("Tasks cleared", "", 0, state.TasksCompleted)},
		},
	}
	for i, task := range state.Tasks {
		panel.Items = append(panel.Items, InteractivePanelItem{
			Key:            fmt.Sprintf("task-%d", i+1),
			Section:        "Open tasks",
			Label:          fmt.Sprintf("Complete task %d", i+1),
			Detail:         normalizeToken(task),
			Status:         "open",
			ApplyInput:     fmt.Sprintf("/tasks done %d", i+1),
			ApplyMode:      PanelApplySubmit,
			PreviewIntents: tasksMutationIntents("Task completed", task, len(state.Tasks)-1, state.TasksCompleted+1),
		})
	}
	return panel
}

func buildEnvInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	envRows := envRows(state)
	panel := InteractivePanel{
		Command:       "env",
		Title:         "env panel: /env",
		Subtitle:      fmt.Sprintf("keys=%d sets=%d", len(envRows), state.EnvSetCount),
		HeaderIntents: envStatusIntents(len(envRows), state.EnvSetCount),
		Items:         []InteractivePanelItem{{Key: "status", Section: "Overview", Label: "Environment status", Detail: "Inspect deterministic environment keys", Status: "status", ApplyInput: "/env status", ApplyMode: PanelApplySubmit, PreviewIntents: envStatusIntents(len(envRows), state.EnvSetCount)}},
	}
	for _, row := range envRows {
		panel.Items = append(panel.Items, InteractivePanelItem{Key: row.key, Section: "Keys", Label: row.key, Detail: row.value, Status: "key", ApplyInput: "/env get " + row.key, ApplyMode: PanelApplySubmit, PreviewIntents: envLookupIntents(row.key, true, row.value)})
	}
	return panel
}

type envRow struct {
	key   string
	value string
}

func envRows(state *RuntimeState) []envRow {
	if state == nil || len(state.ConfigValues) == 0 {
		return nil
	}
	keys := make([]string, 0, len(state.ConfigValues))
	for key := range state.ConfigValues {
		if strings.HasPrefix(key, "env.") {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	out := make([]envRow, 0, len(keys))
	for _, key := range keys {
		plain := strings.TrimPrefix(key, "env.")
		out = append(out, envRow{key: plain, value: state.ConfigValues[key]})
	}
	return out
}

func buildIssueInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	issues := append([]IssueRecord(nil), state.Issues...)
	sort.Slice(issues, func(i, j int) bool { return issues[i].ID < issues[j].ID })
	panel := InteractivePanel{
		Command:       "issue",
		Title:         "issue panel: /issue",
		Subtitle:      fmt.Sprintf("issues=%d provider=%s", len(issues), defaultDash(state.IssueProvider)),
		HeaderIntents: issueStatusIntents(state),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Issue status", Detail: "Inspect counts and provider defaults", Status: "status", ApplyInput: "/issue status", ApplyMode: PanelApplySubmit, PreviewIntents: issueStatusIntents(state)},
			{Key: "list", Section: "Overview", Label: "List issues", Detail: "Render deterministic issue records", Status: "list", ApplyInput: "/issue list", ApplyMode: PanelApplySubmit, PreviewIntents: issueListIntents(issues)},
		},
	}
	for _, issue := range issues {
		action := "/issue close " + issue.ID
		label := "Close " + issue.ID
		status := "open"
		if strings.EqualFold(issue.Status, "closed") {
			action = "/issue open " + issue.ID
			label = "Open " + issue.ID
			status = "closed"
		}
		panel.Items = append(panel.Items, InteractivePanelItem{Key: issue.ID, Section: "Issues", Label: label, Detail: normalizeToken(issue.Title), Status: status, ApplyInput: action, ApplyMode: PanelApplySubmit, PreviewIntents: issueMutationIntents("Issue status", detailRow("ID", normalizeToken(issue.ID), status, "Issue identifier."), detailRow("Title", normalizeToken(issue.Title), "title", "Issue title."), detailRow("Assignee", defaultDash(issue.Assignee), "assignee", "Current assignee."))})
	}
	return panel
}

func buildWorkflowsInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	runs := append([]WorkflowRun(nil), state.WorkflowRuns...)
	sort.Slice(runs, func(i, j int) bool { return runs[i].Name < runs[j].Name })
	panel := InteractivePanel{
		Command:       "workflows",
		Title:         "workflows panel: /workflows",
		Subtitle:      fmt.Sprintf("runs=%d last_action=%s", len(runs), defaultDash(state.WorkflowLastAction)),
		HeaderIntents: workflowsStatusIntents(runs, state.WorkflowLastAction),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Workflow status", Detail: "Inspect run counters and failures", Status: "status", ApplyInput: "/workflows status", ApplyMode: PanelApplySubmit, PreviewIntents: workflowsStatusIntents(runs, state.WorkflowLastAction)},
			{Key: "list", Section: "Overview", Label: "List workflows", Detail: "Show deterministic workflow inventory", Status: "list", ApplyInput: "/workflows list", ApplyMode: PanelApplySubmit, PreviewIntents: workflowsListIntents(runs)},
		},
	}
	for _, run := range runs {
		panel.Items = append(panel.Items, InteractivePanelItem{Key: run.Name, Section: "Workflows", Label: "Rerun " + run.Name, Detail: "status=" + normalizeToken(run.Status), Status: normalizeToken(run.Status), ApplyInput: "/workflows rerun " + run.Name, ApplyMode: PanelApplySubmit, PreviewIntents: workflowsMutationIntents(run.Name, "running", "rerun", run.Provider)})
	}
	return panel
}

func buildProactiveInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	enabled := state.ProactiveEnabled
	panel := InteractivePanel{
		Command:       "proactive",
		Title:         "proactive panel: /proactive",
		Subtitle:      fmt.Sprintf("enabled=%t rules=%d", enabled, len(state.ProactiveRules)),
		HeaderIntents: proactiveStatusIntents(enabled, state.ProactiveRules, state.ProactiveLastAction),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Proactive status", Detail: "Inspect automation policy and rules", Status: statusWord(enabled, "enabled", "disabled"), ApplyInput: "/proactive status", ApplyMode: PanelApplySubmit, PreviewIntents: proactiveStatusIntents(enabled, state.ProactiveRules, state.ProactiveLastAction)},
			{Key: "on", Section: "Mode", Label: "Enable proactive", Detail: "Turn proactive automation on", Status: statusWord(enabled, "current", "mode"), ApplyInput: "/proactive on", ApplyMode: PanelApplySubmit, PreviewIntents: proactiveMutationIntents("Proactive enabled", detailRow("Enabled", "yes", "enabled", "Proactive automation enabled."))},
			{Key: "off", Section: "Mode", Label: "Disable proactive", Detail: "Turn proactive automation off", Status: statusWord(!enabled, "current", "mode"), ApplyInput: "/proactive off", ApplyMode: PanelApplySubmit, PreviewIntents: proactiveMutationIntents("Proactive disabled", detailRow("Enabled", "no", "disabled", "Proactive automation disabled."))},
			{Key: "rules", Section: "Rules", Label: "List rules", Detail: "Show deterministic proactive rule list", Status: statusWord(len(state.ProactiveRules) > 0, "rules", "empty"), ApplyInput: "/proactive rule list", ApplyMode: PanelApplySubmit, PreviewIntents: proactiveStatusIntents(enabled, state.ProactiveRules, state.ProactiveLastAction)},
		},
	}
	return panel
}

func buildAssistantInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	mode := strings.TrimSpace(state.AssistantMode)
	if mode == "" {
		mode = "chat"
	}
	panel := InteractivePanel{
		Command:       "assistant",
		Title:         "assistant panel: /assistant",
		Subtitle:      fmt.Sprintf("mode=%s session=%s", mode, defaultDash(state.AssistantSessionID)),
		HeaderIntents: assistantStatusIntents(mode, state.AssistantSessionID, state.AssistantLastAction),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Assistant status", Detail: "Inspect mode and session id", Status: mode, ApplyInput: "/assistant status", ApplyMode: PanelApplySubmit, PreviewIntents: assistantStatusIntents(mode, state.AssistantSessionID, state.AssistantLastAction)},
			{Key: "chat", Section: "Mode", Label: "Set mode chat", Detail: "Switch assistant mode to chat", Status: statusWord(mode == "chat", "current", "mode"), ApplyInput: "/assistant mode chat", ApplyMode: PanelApplySubmit, PreviewIntents: assistantMutationIntents("Assistant mode", detailRow("Mode", "chat", "mode", "Assistant mode."))},
			{Key: "plan", Section: "Mode", Label: "Set mode plan", Detail: "Switch assistant mode to plan", Status: statusWord(mode == "plan", "current", "mode"), ApplyInput: "/assistant mode plan", ApplyMode: PanelApplySubmit, PreviewIntents: assistantMutationIntents("Assistant mode", detailRow("Mode", "plan", "mode", "Assistant mode."))},
			{Key: "review", Section: "Mode", Label: "Set mode review", Detail: "Switch assistant mode to review", Status: statusWord(mode == "review", "current", "mode"), ApplyInput: "/assistant mode review", ApplyMode: PanelApplySubmit, PreviewIntents: assistantMutationIntents("Assistant mode", detailRow("Mode", "review", "mode", "Assistant mode."))},
			{Key: "reset", Section: "Actions", Label: "Reset assistant", Detail: "Reset mode and session", Status: "reset", ApplyInput: "/assistant reset", ApplyMode: PanelApplySubmit, PreviewIntents: assistantMutationIntents("Assistant reset", detailRow("Mode", "chat", "mode", "Reset to chat mode."), detailRow("Session", "-", "session", "Session cleared."))},
		},
	}
	return panel
}

func buildShareInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	links := append([]ShareLink(nil), state.ShareLinks...)
	sort.Slice(links, func(i, j int) bool { return links[i].ID < links[j].ID })
	panel := InteractivePanel{
		Command:       "share",
		Title:         "share panel: /share",
		Subtitle:      fmt.Sprintf("links=%d last_action=%s", len(links), defaultDash(state.ShareLastAction)),
		HeaderIntents: shareStatusIntents(links, state.ShareLastAction),
		Items: []InteractivePanelItem{
			{Key: "status", Section: "Overview", Label: "Share status", Detail: "Inspect active and revoked links", Status: "status", ApplyInput: "/share status", ApplyMode: PanelApplySubmit, PreviewIntents: shareStatusIntents(links, state.ShareLastAction)},
			{Key: "list", Section: "Overview", Label: "List share links", Detail: "Render deterministic share link table", Status: "list", ApplyInput: "/share list", ApplyMode: PanelApplySubmit, PreviewIntents: shareListIntents(links)},
			{Key: "create", Section: "Actions", Label: "Create share link", Detail: "Create a private session share link", Status: "create", ApplyInput: "/share create", ApplyMode: PanelApplySubmit, PreviewIntents: shareMutationIntents("Share created", detailRow("Scope", "session", "scope", "Default share scope."), detailRow("Visibility", "private", "visibility", "Default visibility."))},
		},
	}
	for _, link := range links {
		if link.Revoked {
			continue
		}
		panel.Items = append(panel.Items, InteractivePanelItem{Key: link.ID, Section: "Active links", Label: "Revoke " + link.ID, Detail: normalizeToken(link.URL), Status: "active", ApplyInput: "/share revoke " + link.ID, ApplyMode: PanelApplySubmit, PreviewIntents: shareMutationIntents("Share revoked", detailRow("ID", normalizeToken(link.ID), "share", "Share link id."), detailRow("URL", normalizeToken(link.URL), "url", "Share URL."))})
	}
	return panel
}
