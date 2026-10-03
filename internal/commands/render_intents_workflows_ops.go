package commands

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

func tasksStatusIntents(tasks []string, completed int) []types.RenderIntent {
	remaining := len(tasks)
	return []types.RenderIntent{
		summaryCardIntent(
			"Tasks",
			"Deterministic local task queue.",
			field("Open", itoa(remaining)),
			field("Completed", itoa(completed)),
		),
		actionHintsIntent("Task actions", hint("List", "/tasks list"), hint("Add", "/tasks add <text>"), hint("Done", "/tasks done <index>"), hint("Clear", "/tasks clear")),
	}
}

func tasksListIntents(tasks []string, completed int) []types.RenderIntent {
	rows := make([]types.RenderTableRow, 0, len(tasks))
	for i, task := range tasks {
		rows = append(rows, tableRow(itoa(i+1), normalizeToken(task)))
	}
	return []types.RenderIntent{
		summaryCardIntent("Task list", "Open tasks for this session.", field("Open", itoa(len(tasks))), field("Completed", itoa(completed))),
		tableIntent("Tasks", "Ordered by deterministic insertion index.", []string{"#", "Task"}, rows...),
	}
}

func tasksMutationIntents(title, task string, count, completed int) []types.RenderIntent {
	rows := []types.RenderDetailRow{
		detailRow("Task", defaultDash(task), "state", "Affected task."),
		detailRow("Open", itoa(count), "count", "Open tasks remaining."),
		detailRow("Completed", itoa(completed), "count", "Total completed tasks."),
	}
	return []types.RenderIntent{detailRowsIntent(strings.TrimSpace(title), "Task mutation result.", rows...)}
}

func envStatusIntents(keys, sets int) []types.RenderIntent {
	return []types.RenderIntent{
		summaryCardIntent("Environment", "Deterministic local environment view.", field("Keys", itoa(keys)), field("Set ops", itoa(sets))),
		actionHintsIntent("Environment actions", hint("Status", "/env status"), hint("Set", "/env set <key> <value>"), hint("Get", "/env get <key>")),
	}
}

func envLookupIntents(key string, found bool, value string) []types.RenderIntent {
	status := "missing"
	if found {
		status = "found"
	}
	return []types.RenderIntent{detailRowsIntent("Environment lookup", "Requested environment key.", detailRow("Key", normalizeToken(key), status, "Requested key."), detailRow("Value", defaultDash(value), status, "Stored value when present."))}
}

func envSetIntents(key, value string, sets int) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Environment updated", "Stored deterministic environment value.", detailRow("Key", normalizeToken(key), "set", "Updated key."), detailRow("Value", normalizeToken(value), "set", "Stored value."), detailRow("Set ops", itoa(sets), "count", "Number of set mutations."))}
}

func issueStatusIntents(state *RuntimeState) []types.RenderIntent {
	if state == nil {
		state = &RuntimeState{}
	}
	openCount := 0
	for _, issue := range state.Issues {
		if strings.EqualFold(issue.Status, "open") {
			openCount++
		}
	}
	provider := normalizeToken(state.IssueProvider)
	if provider == "-" {
		provider = normalizeToken(state.ProviderName)
	}
	return []types.RenderIntent{
		summaryCardIntent("Issues", "Provider-agnostic issue lifecycle.", field("Provider", provider), field("Total", itoa(len(state.Issues))), field("Open", itoa(openCount)), field("Closed", itoa(len(state.Issues)-openCount))),
		actionHintsIntent("Issue actions", hint("List", "/issue list"), hint("Create", "/issue create <title>"), hint("Assign", "/issue assign <id> <assignee>")),
	}
}

func issueListIntents(issues []IssueRecord) []types.RenderIntent {
	rows := make([]types.RenderTableRow, 0, len(issues))
	for _, issue := range issues {
		rows = append(rows, tableRow(normalizeToken(issue.ID), normalizeToken(issue.Status), normalizeToken(issue.Assignee), normalizeToken(issue.Title)))
	}
	return []types.RenderIntent{tableIntent("Issue list", "Provider-agnostic issue records.", []string{"ID", "Status", "Assignee", "Title"}, rows...)}
}

func issueMutationIntents(title string, rows ...types.RenderDetailRow) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent(strings.TrimSpace(title), "Issue mutation result.", rows...)}
}

func workflowsStatusIntents(runs []WorkflowRun, lastAction string) []types.RenderIntent {
	running := 0
	failed := 0
	for _, run := range runs {
		switch strings.ToLower(strings.TrimSpace(run.Status)) {
		case "running":
			running++
		case "failed":
			failed++
		}
	}
	return []types.RenderIntent{
		summaryCardIntent("Workflows", "Provider-agnostic workflow runs.", field("Runs", itoa(len(runs))), field("Running", itoa(running)), field("Failed", itoa(failed)), field("Last action", defaultDash(lastAction))),
		actionHintsIntent("Workflow actions", hint("List", "/workflows list"), hint("Run", "/workflows run <name>"), hint("Rerun", "/workflows rerun <name>")),
	}
}

func workflowsListIntents(runs []WorkflowRun) []types.RenderIntent {
	rows := make([]types.RenderTableRow, 0, len(runs))
	for _, run := range runs {
		rows = append(rows, tableRow(normalizeToken(run.Name), normalizeToken(run.Status), normalizeToken(run.Provider), normalizeToken(run.LastRun)))
	}
	return []types.RenderIntent{tableIntent("Workflow list", "Current deterministic workflow records.", []string{"Name", "Status", "Provider", "Last run"}, rows...)}
}

func workflowsMutationIntents(name, status, reason, provider string) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent("Workflow updated", "Workflow state mutation result.", detailRow("Name", normalizeToken(name), "workflow", "Workflow identifier."), detailRow("Status", normalizeToken(status), normalizeToken(status), "New workflow status."), detailRow("Reason", defaultDash(reason), "detail", "Mutation reason when provided."), detailRow("Provider", defaultDash(provider), "provider", "Workflow provider."))}
}

func proactiveStatusIntents(enabled bool, rules []string, lastAction string) []types.RenderIntent {
	sorted := uniqueSortedStrings(append([]string(nil), rules...))
	groups := []types.RenderGroup{group("Rules", "Configured proactive rules.", sorted...)}
	return []types.RenderIntent{
		summaryCardIntent("Proactive", "Automation policy and triggers.", field("Enabled", boolState(enabled, "yes", "no")), field("Rules", itoa(len(sorted))), field("Last action", defaultDash(lastAction))),
		groupedListIntent("Rule groups", "Rule inventory by deterministic sort order.", groups...),
	}
}

func proactiveMutationIntents(title string, rows ...types.RenderDetailRow) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent(strings.TrimSpace(title), "Proactive policy mutation.", rows...)}
}

func assistantStatusIntents(mode, sessionID, lastAction string) []types.RenderIntent {
	if strings.TrimSpace(mode) == "" {
		mode = "chat"
	}
	return []types.RenderIntent{
		summaryCardIntent("Assistant", "Assistant mode and session controls.", field("Mode", normalizeToken(mode)), field("Session", defaultDash(sessionID)), field("Last action", defaultDash(lastAction))),
		actionHintsIntent("Assistant actions", hint("Mode", "/assistant mode <chat|plan|review>"), hint("Session", "/assistant session [id]"), hint("Reset", "/assistant reset")),
	}
}

func assistantMutationIntents(title string, rows ...types.RenderDetailRow) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent(strings.TrimSpace(title), "Assistant mutation result.", rows...)}
}

func shareStatusIntents(links []ShareLink, lastAction string) []types.RenderIntent {
	active := 0
	for _, link := range links {
		if !link.Revoked {
			active++
		}
	}
	return []types.RenderIntent{
		summaryCardIntent("Share links", "Deterministic share link inventory.", field("Total", itoa(len(links))), field("Active", itoa(active)), field("Revoked", itoa(len(links)-active)), field("Last action", defaultDash(lastAction))),
		actionHintsIntent("Share actions", hint("List", "/share list"), hint("Create", "/share create [scope] [private|public]"), hint("Revoke", "/share revoke <id>")),
	}
}

func shareListIntents(links []ShareLink) []types.RenderIntent {
	sorted := append([]ShareLink(nil), links...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].ID < sorted[j].ID })
	rows := make([]types.RenderTableRow, 0, len(sorted))
	for _, link := range sorted {
		rows = append(rows, tableRow(normalizeToken(link.ID), normalizeToken(link.Scope), normalizeToken(link.Visibility), boolState(link.Revoked, "yes", "no"), normalizeToken(link.URL)))
	}
	return []types.RenderIntent{tableIntent("Share list", "Published share links for this session.", []string{"ID", "Scope", "Visibility", "Revoked", "URL"}, rows...)}
}

func shareMutationIntents(title string, rows ...types.RenderDetailRow) []types.RenderIntent {
	return []types.RenderIntent{detailRowsIntent(strings.TrimSpace(title), "Share link mutation result.", rows...)}
}

func itoa(v int) string { return fmt.Sprintf("%d", v) }
