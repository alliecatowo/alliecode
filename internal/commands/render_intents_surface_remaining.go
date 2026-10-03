package commands

import (
	"fmt"
	"sort"

	"github.com/alliecatowo/alliecode/internal/types"
)

func terminalSetupDetectIntents(profile, source string, supported bool, hints []string) []types.RenderIntent {
	opts := make([]types.RenderOption, 0, len(hints))
	for _, h := range hints {
		opts = append(opts, option(h, "Terminal setup hint", "hint", "", false))
	}
	return []types.RenderIntent{
		summaryCardIntent("Terminal setup detect", "Detected terminal profile and support state.",
			field("Profile", defaultDash(profile)),
			field("Source", defaultDash(source)),
			field("Supported", boolState(supported, "yes", "no")),
		),
		optionListIntent("Hints", "Recommended setup steps.", opts...),
	}
}

func terminalSetupStatusIntents(state *RuntimeState) []types.RenderIntent {
	opts := make([]types.RenderOption, 0, len(state.TerminalSetupHints))
	for _, h := range state.TerminalSetupHints {
		opts = append(opts, option(h, "Saved setup hint", "hint", "", false))
	}
	return []types.RenderIntent{
		summaryCardIntent("Terminal setup status", "Terminal setup counters and detected profile.",
			field("Configured", boolState(state.TerminalConfigured, "yes", "no")),
			field("Count", fmt.Sprintf("%d", state.TerminalSetupCount)),
			field("Profile", defaultDash(state.TerminalProfile)),
			field("Detected", defaultDash(state.TerminalDetectedProfile)),
			field("Source", defaultDash(state.TerminalDetectSource)),
		),
		optionListIntent("Hints", "Stored setup hints.", opts...),
	}
}

func terminalSetupApplyIntents(profile string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Terminal setup applied", "Recorded terminal setup profile.",
		detailRow("Profile", defaultDash(profile), "applied", "Selected terminal profile."),
		detailRow("Configured", "yes", "ok", "Terminal integration is marked configured."),
		detailRow("Count", fmt.Sprintf("%d", count), "info", "Apply invocation count."),
	)}
}

func releaseNotesStatusIntents(seen int, lastVersion string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Release notes status", "Release notes counters and last viewed version.", detailRow("Seen", fmt.Sprintf("%d", seen), "info", "View count."), detailRow("Last version", defaultDash(lastVersion), "info", "Last viewed version."))}
}

func releaseNotesListIntents(versions []string, seen int) []types.RenderIntent {
	rows := make([]types.RenderTableRow, 0, len(versions))
	for _, version := range versions {
		rows = append(rows, tableRow(version))
	}
	return []types.RenderIntent{
		summaryCardIntent("Release notes list", "Known release versions.", field("Seen", fmt.Sprintf("%d", seen))),
		tableIntent("Versions", "Deterministic release snapshot.", []string{"Version"}, rows...),
	}
}

func releaseNotesLatestIntents(version string, seen int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Release notes latest", "Latest deterministic release notes entry.",
		detailRow("Version", defaultDash(version), "latest", "Most recent version."),
		detailRow("Seen", fmt.Sprintf("%d", seen), "info", "View count."),
	)}
}

func installFlowIntents(title string, rows ...types.RenderDetailRow) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent(title, "Install flow state.", rows...)}
}

func feedbackStatusIntents(count int, last string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Feedback status", "Feedback submission counters.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Submission count."), detailRow("Last", defaultDash(last), "info", "Last feedback message."))}
}

func feedbackSubmitIntents(message string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Feedback submitted", "Recorded deterministic feedback payload.", detailRow("Message", defaultDash(message), "submitted", "Submitted message."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Submission count."))}
}

func hooksStatusIntents(pre, post bool) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Hooks status", "Pre and post hook toggles.", detailRow("Pre", boolState(pre, "on", "off"), "state", "Pre hook state."), detailRow("Post", boolState(post, "on", "off"), "state", "Post hook state."))}
}

func hooksSetIntents(action, target string, pre, post bool) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Hooks updated", "Hook toggle mutation result.", detailRow("Action", defaultDash(action), "info", "Mutation action."), detailRow("Target", defaultDash(target), "info", "Target hook."), detailRow("Pre", boolState(pre, "on", "off"), "state", "Pre-hook state."), detailRow("Post", boolState(post, "on", "off"), "state", "Post-hook state."))}
}

func advisorStatusIntents(model string, active bool, updates int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Advisor status", "Advisor model selection.", detailRow("Model", defaultDash(model), "info", "Advisor model."), detailRow("Active", boolState(active, "yes", "no"), "state", "Whether advisor is active."), detailRow("Updates", fmt.Sprintf("%d", updates), "info", "Mutation count."))}
}

func advisorSetIntents(model string, active bool, updates int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Advisor updated", "Advisor setting mutation result.", detailRow("Model", defaultDash(model), "info", "Advisor model."), detailRow("Active", boolState(active, "yes", "no"), "state", "Whether advisor is active."), detailRow("Updates", fmt.Sprintf("%d", updates), "info", "Mutation count."))}
}

func btwIntents(question string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("BTW queued", "Queued side question.", detailRow("Question", defaultDash(question), "queued", "Submitted side question."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Invocation count."))}
}

func chromeStatusIntents(defaultEnabled, installed, connected bool, actions int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Chrome status", "Claude in Chrome integration status.", detailRow("Default enabled", boolState(defaultEnabled, "yes", "no"), "state", "Default routing state."), detailRow("Extension installed", boolState(installed, "yes", "no"), "state", "Extension state."), detailRow("Connected", boolState(connected, "yes", "no"), "state", "Connection state."), detailRow("Actions", fmt.Sprintf("%d", actions), "info", "Action count."))}
}

func chromeActionIntents(title string, defaultEnabled, installed, connected bool, actions int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent(title, "Chrome integration action result.", detailRow("Default enabled", boolState(defaultEnabled, "yes", "no"), "state", "Default routing state."), detailRow("Extension installed", boolState(installed, "yes", "no"), "state", "Extension state."), detailRow("Connected", boolState(connected, "yes", "no"), "state", "Connection state."), detailRow("Actions", fmt.Sprintf("%d", actions), "info", "Action count."))}
}

func colorStatusIntents(color string, setCount int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Color status", "Session color status.", detailRow("Color", defaultDash(color), "info", "Current session color."), detailRow("Set count", fmt.Sprintf("%d", setCount), "info", "Mutation count."))}
}

func colorSetIntents(color string, setCount int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Color updated", "Session color mutation result.", detailRow("Color", defaultDash(color), "set", "Selected color token."), detailRow("Set count", fmt.Sprintf("%d", setCount), "info", "Mutation count."))}
}

func desktopStatusIntents(count int, lastTarget string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Desktop status", "Desktop handoff status.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Handoff count."), detailRow("Last target", defaultDash(lastTarget), "info", "Most recent handoff target."))}
}

func desktopHandoffIntents(target string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Desktop handoff", "Requested desktop handoff.", detailRow("Target", defaultDash(target), "queued", "Target app."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Invocation count."))}
}

func mobileStatusIntents(platform string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Mobile status", "Mobile QR status.", detailRow("Platform", defaultDash(platform), "info", "Current platform."), detailRow("Count", fmt.Sprintf("%d", count), "info", "QR generation count."))}
}

func mobileQRIntents(platform string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Mobile QR", "Generated mobile QR state.", detailRow("Platform", defaultDash(platform), "generated", "Selected platform."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Generation count."))}
}

func fastStatusIntents(enabled bool, toggles int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Fast mode status", "Fast mode toggle state.", detailRow("Enabled", boolState(enabled, "yes", "no"), "state", "Current fast mode."), detailRow("Toggles", fmt.Sprintf("%d", toggles), "info", "Toggle count."))}
}

func fastSetIntents(enabled bool, toggles int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Fast mode updated", "Fast mode mutation result.", detailRow("Enabled", boolState(enabled, "yes", "no"), statusWord(enabled, "enabled", "disabled"), "Current fast mode."), detailRow("Toggles", fmt.Sprintf("%d", toggles), "info", "Toggle count."))}
}

func effortStatusIntents(value string, updates int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Effort status", "Effort preference status.", detailRow("Value", defaultDash(value), "info", "Current effort value."), detailRow("Updates", fmt.Sprintf("%d", updates), "info", "Mutation count."))}
}

func effortSetIntents(value string, updates int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Effort updated", "Effort preference mutation result.", detailRow("Value", defaultDash(value), "set", "Current effort value."), detailRow("Updates", fmt.Sprintf("%d", updates), "info", "Mutation count."))}
}

func tasksListStateIntents(tasks []string, completed int) []types.RenderIntent {
	rows := make([]types.RenderDetailRow, 0, len(tasks)+2)
	rows = append(rows, detailRow("Open", fmt.Sprintf("%d", len(tasks)), "info", "Open task count."))
	rows = append(rows, detailRow("Completed", fmt.Sprintf("%d", completed), "info", "Completed task count."))
	for i, task := range tasks {
		rows = append(rows, detailRow(fmt.Sprintf("Task %d", i+1), task, "open", "Queued task."))
	}
	return []types.RenderIntent{detailRowsIntent("Tasks list", "Current deterministic task list.", rows...)}
}

func reloadPluginsIntents(pendingBefore bool, installed, enabled, marketplaces, reloadCount int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Reload plugins", "Plugin reload operation result.", detailRow("Pending before", boolState(pendingBefore, "yes", "no"), "state", "Reload pending before command."), detailRow("Installed", fmt.Sprintf("%d", installed), "info", "Installed plugins count."), detailRow("Enabled", fmt.Sprintf("%d", enabled), "info", "Enabled plugins count."), detailRow("Marketplaces", fmt.Sprintf("%d", marketplaces), "info", "Configured marketplaces count."), detailRow("Reload count", fmt.Sprintf("%d", reloadCount), "info", "Reload invocation count."))}
}

func exportResultIntents(path string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Export result", "Conversation export destination.", detailRow("Path", defaultDash(path), "written", "Export path."), detailRow("Format", "txt", "info", "Export format."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Export count."))}
}

func extraUsageStatusIntents(enabled bool, requests int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Extra usage status", "Extra usage eligibility and requests.", detailRow("Enabled", boolState(enabled, "yes", "no"), "state", "Enablement state."), detailRow("Requests", fmt.Sprintf("%d", requests), "info", "Request count."))}
}

func extraUsageSetIntents(enabled bool, requests int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Extra usage updated", "Extra usage flag updated.", detailRow("Enabled", boolState(enabled, "yes", "no"), "state", "Enablement state."), detailRow("Requests", fmt.Sprintf("%d", requests), "info", "Request count."))}
}

func extraUsageRequestIntents(enabled bool, requests int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Extra usage requested", "Recorded extra usage request.", detailRow("Enabled", boolState(enabled, "yes", "no"), "state", "Current enablement."), detailRow("Requests", fmt.Sprintf("%d", requests), "info", "Request count."))}
}

func rateLimitOptionsIntents(options []string, prompts int, lastAction string) []types.RenderIntent {
	opts := make([]types.RenderOption, 0, len(options))
	for _, o := range options {
		opts = append(opts, option(o, "Available rate-limit action", "option", "/rate-limit-options "+o, false))
	}
	return []types.RenderIntent{detailRowsIntent("Rate limit options", "Rate-limit decision helper state.", detailRow("Prompts", fmt.Sprintf("%d", prompts), "info", "Rate-limit prompt count."), detailRow("Last action", defaultDash(lastAction), "info", "Most recent selected action.")), optionListIntent("Options", "Actions available after limit trigger.", opts...)}
}

func rateLimitActionIntents(action string, prompts int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Rate limit action", "Recorded selected rate-limit action.", detailRow("Action", defaultDash(action), "selected", "Chosen action."), detailRow("Prompts", fmt.Sprintf("%d", prompts), "info", "Prompt count."))}
}

func prCommentsStatusIntents(fetches int, lastRef string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("PR comments status", "PR comments fetch status.", detailRow("Fetches", fmt.Sprintf("%d", fetches), "info", "Fetch count."), detailRow("Last ref", defaultDash(lastRef), "info", "Last fetched ref."))}
}

func prCommentsSummaryIntents(ref, source string, fetches int, threads []commentThread) []types.RenderIntent {
	rows := make([]types.RenderTableRow, 0, len(threads))
	for _, thread := range threads {
		rows = append(rows, tableRow(thread.Path, fmt.Sprintf("%d", thread.Line), fmt.Sprintf("%d", len(thread.Comments))))
	}
	return []types.RenderIntent{summaryCardIntent("PR comments summary", "Thread counts and source metadata.", field("Ref", defaultDash(ref)), field("Source", defaultDash(source)), field("Fetches", fmt.Sprintf("%d", fetches)), field("Threads", fmt.Sprintf("%d", len(threads)))), tableIntent("Threads", "Comment threads grouped by path and line.", []string{"Path", "Line", "Comments"}, rows...)}
}

func webSetupStatusIntents(connected bool, url string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Web setup status", "Web setup connection status.", detailRow("Connected", boolState(connected, "yes", "no"), "state", "Connection state."), detailRow("URL", defaultDash(url), "info", "Session URL."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Mutation count."))}
}

func webSetupConnectIntents(connected bool, url string, count int) []types.RenderIntent {
	status := "connected"
	if !connected {
		status = "disconnected"
	}
	return []types.RenderIntent{detailRowsIntent("Web setup updated", "Web setup connection action result.", detailRow("Connected", boolState(connected, "yes", "no"), status, "Current connection state."), detailRow("URL", defaultDash(url), "info", "Web session URL."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Mutation count."))}
}

func bridgeKickStatusIntents(count int, lastAction string, lastCode int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Bridge kick status", "Bridge fault-injection status.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Invocation count."), detailRow("Last action", defaultDash(lastAction), "info", "Most recent action."), detailRow("Last code", fmt.Sprintf("%d", lastCode), "info", "Most recent code."))}
}

func bridgeKickApplyIntents(action string, code, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Bridge kick apply", "Applied deterministic bridge kick action.", detailRow("Action", defaultDash(action), "applied", "Applied action."), detailRow("Code", fmt.Sprintf("%d", code), "info", "Associated code."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Invocation count."))}
}

func briefStatusIntents(enabled bool, toggles int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Brief mode status", "Brief-only mode status.", detailRow("Enabled", boolState(enabled, "yes", "no"), "state", "Current mode."), detailRow("Toggles", fmt.Sprintf("%d", toggles), "info", "Toggle count."))}
}

func briefSetIntents(enabled bool, toggles int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Brief mode updated", "Brief-only mode changed.", detailRow("Enabled", boolState(enabled, "yes", "no"), "state", "Current brief mode."), detailRow("Toggles", fmt.Sprintf("%d", toggles), "info", "Toggle count."))}
}

func initVerifiersStatusIntents(count int, lastName string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Init verifiers status", "Verifier scaffold status.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Creation count."), detailRow("Last name", defaultDash(lastName), "info", "Last verifier name."))}
}

func initVerifiersCreatedIntents(name string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Init verifiers created", "Verifier scaffold created.", detailRow("Name", defaultDash(name), "created", "Verifier name."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Creation count."))}
}

func insightsStatusIntents(count int, lastScope string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Insights status", "Insights generation status.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Report count."), detailRow("Last scope", defaultDash(lastScope), "info", "Last report scope."))}
}

func insightsReportIntents(scope string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Insights report", "Generated deterministic insights report.", detailRow("Scope", defaultDash(scope), "generated", "Report scope."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Generation count."))}
}

func passesStatusIntents(visits, remaining int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Passes status", "Guest pass inventory status.", detailRow("Visits", fmt.Sprintf("%d", visits), "info", "Visit count."), detailRow("Remaining", fmt.Sprintf("%d", remaining), "info", "Remaining passes."))}
}

func passesClaimIntents(claimed, remaining, visits int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Passes claimed", "Guest pass claim result.", detailRow("Claimed", fmt.Sprintf("%d", claimed), "claimed", "Claimed pass count."), detailRow("Remaining", fmt.Sprintf("%d", remaining), "info", "Remaining passes."), detailRow("Visits", fmt.Sprintf("%d", visits), "info", "Visit count."))}
}

func renameStatusIntents(count int, name string, historyMatches int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Rename status", "Conversation title status.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Rename count."), detailRow("Name", defaultDash(name), "info", "Current title."), detailRow("History matches", fmt.Sprintf("%d", historyMatches), "info", "Matching history rows."))}
}

func renameSetIntents(name string, count, updated int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Rename updated", "Conversation title updated.", detailRow("Name", defaultDash(name), "set", "Updated title."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Rename count."), detailRow("History updated", fmt.Sprintf("%d", updated), "info", "Updated history rows."))}
}

func stickersStatusIntents(count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Stickers status", "Sticker order status.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Order count."))}
}

func stickersOrderIntents(count int, url string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Stickers ordered", "Sticker order request state.", detailRow("Count", fmt.Sprintf("%d", count), "ordered", "Order count."), detailRow("URL", defaultDash(url), "info", "Order URL."))}
}

func thinkbackStatusIntents(count int, lastAction string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Thinkback status", "Thinkback status.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Invocation count."), detailRow("Last action", defaultDash(lastAction), "info", "Most recent action."))}
}

func thinkbackActionIntents(action string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Thinkback action", "Thinkback action result.", detailRow("Action", defaultDash(action), "executed", "Executed action."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Invocation count."))}
}

func thinkbackPlayStatusIntents(count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Thinkback play status", "Thinkback playback status.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Playback count."))}
}

func thinkbackPlayRunIntents(count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Thinkback play", "Triggered thinkback playback.", detailRow("Count", fmt.Sprintf("%d", count), "played", "Playback count."))}
}

func teleportStatusIntents(count int, lastTarget string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Teleport status", "Teleport handoff status.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Handoff count."), detailRow("Last target", defaultDash(lastTarget), "info", "Most recent target."))}
}

func teleportSetIntents(target string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Teleport target set", "Updated deterministic teleport target.", detailRow("Target", defaultDash(target), "set", "Selected target."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Mutation count."))}
}

func summaryStatusIntents(count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Summary status", "Summary refresh status.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Refresh count."))}
}

func summaryRefreshIntents(count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Summary refreshed", "Triggered summary refresh.", detailRow("Count", fmt.Sprintf("%d", count), "refreshed", "Refresh count."))}
}

func resetLimitsStatusIntents(count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Reset limits status", "Reset-limits status.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Reset invocation count."))}
}

func resetLimitsApplyIntents(count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Limits reset", "Reset deterministic usage counters.", detailRow("Count", fmt.Sprintf("%d", count), "reset", "Reset invocation count."))}
}

func oauthRefreshStatusIntents(states map[string]OAuthRefreshState, lastAction string) []types.RenderIntent {
	providers := make([]string, 0, len(states))
	for provider := range states {
		providers = append(providers, provider)
	}
	sort.Strings(providers)
	rows := make([]types.RenderTableRow, 0, len(providers))
	for _, provider := range providers {
		state := states[provider]
		rows = append(rows, tableRow(provider, boolState(state.Refreshed, "yes", "no"), fmt.Sprintf("%d", state.ExpiresIn), defaultDash(state.Error)))
	}
	return []types.RenderIntent{detailRowsIntent("OAuth refresh status", "OAuth refresh status by provider.", detailRow("Count", fmt.Sprintf("%d", len(providers)), "info", "Provider count."), detailRow("Last action", defaultDash(lastAction), "info", "Most recent action.")), tableIntent("Providers", "Tracked OAuth provider refresh states.", []string{"Provider", "Refreshed", "Expires in", "Error"}, rows...)}
}

func oauthRefreshResultIntents(provider string, refreshed bool, expiresIn int, tokenPrefix, reason string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("OAuth refresh result", "OAuth refresh action result.", detailRow("Provider", defaultDash(provider), "info", "Target provider."), detailRow("Refreshed", boolState(refreshed, "yes", "no"), "state", "Refresh status."), detailRow("Expires in", fmt.Sprintf("%d", expiresIn), "info", "Token expiry window in seconds."), detailRow("Token prefix", defaultDash(tokenPrefix), "info", "Token prefix preview."), detailRow("Error", defaultDash(reason), "info", "Error detail when refresh fails."))}
}

func oauthRefreshClearIntents(provider string, removed bool) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("OAuth refresh cleared", "Removed tracked OAuth refresh state.", detailRow("Provider", defaultDash(provider), "info", "Target provider."), detailRow("Removed", boolState(removed, "yes", "no"), "state", "Whether state existed."))}
}

func antTraceStatusIntents(enabled bool, count, marks int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Ant trace status", "Ant-trace status.", detailRow("Enabled", boolState(enabled, "yes", "no"), "state", "Current state."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Invocation count."), detailRow("Marks", fmt.Sprintf("%d", marks), "info", "Mark count."))}
}

func antTraceSetIntents(enabled bool, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Ant trace updated", "Ant-trace mode updated.", detailRow("Enabled", boolState(enabled, "yes", "no"), "state", "Current ant-trace state."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Mutation count."))}
}

func antTraceMarkIntents(label string, marks, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Ant trace mark", "Recorded ant-trace mark.", detailRow("Label", defaultDash(label), "marked", "Recorded mark label."), detailRow("Marks", fmt.Sprintf("%d", marks), "info", "Total marks."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Mutation count."))}
}

func antTraceListIntents(marks []string) []types.RenderIntent {
	rows := make([]types.RenderTableRow, 0, len(marks))
	for i, mark := range marks {
		rows = append(rows, tableRow(fmt.Sprintf("%d", i+1), mark))
	}
	return []types.RenderIntent{tableIntent("Ant trace marks", "Current ant-trace marks.", []string{"#", "Mark"}, rows...)}
}

func antTraceClearIntents(count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Ant trace cleared", "Cleared ant-trace marks.", detailRow("Count", fmt.Sprintf("%d", count), "cleared", "Mutation count."))}
}

func autofixPRStatusIntents(count int, lastAction, lastRef string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Autofix PR status", "Autofix PR workflow status.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Invocation count."), detailRow("Last action", defaultDash(lastAction), "info", "Most recent action."), detailRow("Last ref", defaultDash(lastRef), "info", "Most recent ref."))}
}

func autofixPRActionIntents(action, ref string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Autofix PR action", "Autofix PR action result.", detailRow("Action", defaultDash(action), "executed", "Executed action."), detailRow("Ref", defaultDash(ref), "info", "Target ref."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Invocation count."))}
}

func autofixPRCancelIntents(count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Autofix PR cancelled", "Cancelled queued autofix PR action.", detailRow("Count", fmt.Sprintf("%d", count), "cancelled", "Invocation count."))}
}

func backfillStatusIntents(runs, lastCount int, lastAction string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Backfill sessions status", "Backfill run status.", detailRow("Runs", fmt.Sprintf("%d", runs), "info", "Run count."), detailRow("Last count", fmt.Sprintf("%d", lastCount), "info", "Most recent batch size."), detailRow("Last action", defaultDash(lastAction), "info", "Most recent action."))}
}

func backfillRunIntents(action string, count, runs int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Backfill run", "Backfill run action result.", detailRow("Action", defaultDash(action), "executed", "Run action."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Requested count."), detailRow("Runs", fmt.Sprintf("%d", runs), "info", "Total runs."))}
}

func breakCacheStatusIntents(total int, lastScope string, scopes map[string]int) []types.RenderIntent {
	rows := []types.RenderTableRow{
		tableRow("all", fmt.Sprintf("%d", scopes["all"])),
		tableRow("models", fmt.Sprintf("%d", scopes["models"])),
		tableRow("history", fmt.Sprintf("%d", scopes["history"])),
		tableRow("tools", fmt.Sprintf("%d", scopes["tools"])),
	}
	return []types.RenderIntent{detailRowsIntent("Break cache status", "Cache-break status.", detailRow("Count", fmt.Sprintf("%d", total), "info", "Total invocations."), detailRow("Last scope", defaultDash(lastScope), "info", "Most recent scope.")), tableIntent("Scopes", "Cache-break counts by scope.", []string{"Scope", "Count"}, rows...)}
}

func breakCacheApplyIntents(scope string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Break cache applied", "Cache scope invalidation recorded.", detailRow("Scope", defaultDash(scope), "applied", "Invalidated scope."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Invocation count."))}
}

func bughunterStatusIntents(count int, lastScope string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Bughunter status", "Bughunter run status.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Run count."), detailRow("Last scope", defaultDash(lastScope), "info", "Most recent scope."))}
}

func bughunterRunIntents(scope string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Bughunter run", "Bughunter run recorded.", detailRow("Scope", defaultDash(scope), "run", "Run scope."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Run count."))}
}

func ctxVizStatusIntents(count int, lastAction string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Context viz status", "Context visualization status.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Invocation count."), detailRow("Last action", defaultDash(lastAction), "info", "Most recent action."))}
}

func ctxVizRenderIntents(focus string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Context viz render", "Context visualization refreshed.", detailRow("Focus", defaultDash(focus), "rendered", "Render focus."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Invocation count."))}
}

func ctxVizClearIntents(count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Context viz clear", "Context visualization state cleared.", detailRow("Count", fmt.Sprintf("%d", count), "cleared", "Invocation count."))}
}

func debugToolCallStatusIntents(count int, lastTool string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Debug tool call status", "Debug tool-call status.", detailRow("Count", fmt.Sprintf("%d", count), "info", "Invocation count."), detailRow("Last tool", defaultDash(lastTool), "info", "Most recent tool."))}
}

func debugToolCallLogIntents(tool string, count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Debug tool call log", "Logged tool-call debug marker.", detailRow("Tool", defaultDash(tool), "logged", "Logged tool."), detailRow("Count", fmt.Sprintf("%d", count), "info", "Invocation count."))}
}

func debugToolCallClearIntents(count int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Debug tool call clear", "Cleared tool-call debug marker.", detailRow("Count", fmt.Sprintf("%d", count), "cleared", "Invocation count."))}
}

func sandboxStatusIntents(settings permissionsSandboxSummary, diagnostics sandboxDiagnostics) []types.RenderIntent {
	return []types.RenderIntent{summaryCardIntent("Sandbox", "Sandbox policy and runtime availability status.",
		field("Mode", defaultDash(settings.Mode)),
		field("Workspace locked", boolState(settings.WorkspaceLocked, "yes", "no")),
		field("Excluded count", fmt.Sprintf("%d", settings.ExcludedCount)),
		field("Shell", defaultDash(diagnostics.ShellStatus)),
		field("Dependency", defaultDash(diagnostics.DependencyStatus)),
	)}
}

type permissionsSandboxSummary struct {
	Mode            string
	WorkspaceLocked bool
	ExcludedCount   int
}

func sandboxCheckIntents(status string, checks []sandboxPreflightCheck) []types.RenderIntent {
	diags := make([]types.RenderDiagnostic, 0, len(checks))
	for _, check := range checks {
		diags = append(diags, diagnostic(check.ID, check.Status, check.Detail))
	}
	return []types.RenderIntent{diagnosticsIntent("Sandbox check", status, "Sandbox preflight checks.", diags, nil)}
}

func sandboxExcludeListIntents(patterns []string) []types.RenderIntent {
	rows := make([]types.RenderTableRow, 0, len(patterns))
	for i, pattern := range patterns {
		rows = append(rows, tableRow(fmt.Sprintf("%d", i+1), pattern))
	}
	return []types.RenderIntent{tableIntent("Sandbox excludes", "Excluded sandbox command patterns.", []string{"#", "Pattern"}, rows...)}
}

func commitSuggestIntents(source, branch string, files int, message string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Commit suggest", "Suggested deterministic commit message.", detailRow("Source", defaultDash(source), "info", "Summary source."), detailRow("Branch", defaultDash(branch), "info", "Active branch."), detailRow("Files", fmt.Sprintf("%d", files), "info", "Changed file count."), detailRow("Suggested message", defaultDash(message), "info", "Suggested commit message."))}
}

func commitCheckIntents(ready bool, reason string, files, staged, unstaged, untracked int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Commit check", "Commit readiness summary.", detailRow("Ready", boolState(ready, "yes", "no"), "state", "Commit readiness."), detailRow("Reason", defaultDash(reason), "info", "Readiness reason."), detailRow("Files", fmt.Sprintf("%d", files), "info", "Changed files."), detailRow("Staged", fmt.Sprintf("%d", staged), "info", "Staged files."), detailRow("Unstaged", fmt.Sprintf("%d", unstaged), "info", "Unstaged files."), detailRow("Untracked", fmt.Sprintf("%d", untracked), "info", "Untracked files."))}
}

func commitPushPRDoctorIntents(branch, upstream string, tracked bool, files, staged int, quickFix string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Commit push PR doctor", "Commit/push/PR doctor guidance.", detailRow("Branch", defaultDash(branch), "info", "Active branch."), detailRow("Tracked", boolState(tracked, "yes", "no"), "state", "Tracking status."), detailRow("Upstream", defaultDash(upstream), "info", "Upstream branch."), detailRow("Files", fmt.Sprintf("%d", files), "info", "Changed files."), detailRow("Staged", fmt.Sprintf("%d", staged), "info", "Staged files."), detailRow("Quick fix", defaultDash(quickFix), "hint", "Suggested next command."))}
}
