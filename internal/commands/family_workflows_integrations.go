package commands

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

func executeIssueCommand(cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || equalFoldTrimmed(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /issue [status|list|provider <name>|create <title>|open <id>|close <id>|assign <id> <assignee>|label <id> <label>|unlabel <id> <label>]")
		}
		openCount := 0
		for _, issue := range cmdCtx.State.Issues {
			if strings.EqualFold(issue.Status, "open") {
				openCount++
			}
		}
		provider := normalizeToken(cmdCtx.State.IssueProvider)
		if provider == "-" {
			provider = normalizeToken(cmdCtx.State.ProviderName)
		}
		message := fmt.Sprintf("ISSUE_STATUS\nprovider=%s\ncount=%d\nopen=%d\nclosed=%d\nlast_action=%s", provider, len(cmdCtx.State.Issues), openCount, len(cmdCtx.State.Issues)-openCount, normalizeToken(cmdCtx.State.IssueLastAction))
		return resultWithIntents(message, issueStatusIntents(cmdCtx.State)...), nil
	}

	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "list":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /issue [status|list|provider <name>|create <title>|open <id>|close <id>|assign <id> <assignee>|label <id> <label>|unlabel <id> <label>]")
		}
		issues := append([]IssueRecord(nil), cmdCtx.State.Issues...)
		sort.Slice(issues, func(i, j int) bool { return issues[i].ID < issues[j].ID })
		lines := []string{"ISSUE_LIST", fmt.Sprintf("count=%d", len(issues))}
		for i, issue := range issues {
			idx := i + 1
			lines = append(lines, fmt.Sprintf("issue.%d.id=%s", idx, normalizeToken(issue.ID)))
			lines = append(lines, fmt.Sprintf("issue.%d.title=%s", idx, normalizeToken(issue.Title)))
			lines = append(lines, fmt.Sprintf("issue.%d.status=%s", idx, normalizeToken(issue.Status)))
			lines = append(lines, fmt.Sprintf("issue.%d.provider=%s", idx, normalizeToken(issue.Provider)))
			lines = append(lines, fmt.Sprintf("issue.%d.assignee=%s", idx, normalizeToken(issue.Assignee)))
		}
		return resultWithIntents(strings.Join(lines, "\n"), issueListIntents(issues)...), nil
	case "provider":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: /issue [status|list|provider <name>|create <title>|open <id>|close <id>|assign <id> <assignee>|label <id> <label>|unlabel <id> <label>]")
		}
		provider := strings.ToLower(strings.TrimSpace(inv.Args[1]))
		if provider == "" {
			return Result{}, fmt.Errorf("usage: /issue [status|list|provider <name>|create <title>|open <id>|close <id>|assign <id> <assignee>|label <id> <label>|unlabel <id> <label>]")
		}
		cmdCtx.State.IssueProvider = provider
		cmdCtx.State.IssueLastAction = "provider"
		message := fmt.Sprintf("ISSUE_PROVIDER\nprovider=%s", normalizeToken(provider))
		return resultWithIntents(message, issueMutationIntents("Issue provider", detailRow("Provider", normalizeToken(provider), "provider", "Default provider for issue operations."))...), nil
	case "create":
		if len(inv.Args) < 2 {
			return Result{}, fmt.Errorf("usage: /issue [status|list|provider <name>|create <title>|open <id>|close <id>|assign <id> <assignee>|label <id> <label>|unlabel <id> <label>]")
		}
		title := strings.TrimSpace(strings.Join(inv.Args[1:], " "))
		if title == "" {
			return Result{}, fmt.Errorf("usage: /issue [status|list|provider <name>|create <title>|open <id>|close <id>|assign <id> <assignee>|label <id> <label>|unlabel <id> <label>]")
		}
		if cmdCtx.State.IssueNextID <= 0 {
			cmdCtx.State.IssueNextID = 1
		}
		id := fmt.Sprintf("ISSUE-%d", cmdCtx.State.IssueNextID)
		cmdCtx.State.IssueNextID++
		provider := strings.TrimSpace(cmdCtx.State.IssueProvider)
		if provider == "" {
			provider = strings.TrimSpace(cmdCtx.State.ProviderName)
		}
		if provider == "" {
			provider = "generic"
		}
		issue := IssueRecord{ID: id, Title: title, Status: "open", Provider: provider}
		cmdCtx.State.Issues = append(cmdCtx.State.Issues, issue)
		cmdCtx.State.IssueLastAction = "create"
		message := fmt.Sprintf("ISSUE_CREATE\nid=%s\nstatus=open\nprovider=%s\ntitle=%s\ncount=%d", normalizeToken(id), normalizeToken(provider), normalizeToken(title), len(cmdCtx.State.Issues))
		return resultWithIntents(message, issueMutationIntents("Issue created", detailRow("ID", normalizeToken(id), "open", "Created issue identifier."), detailRow("Provider", normalizeToken(provider), "provider", "Issue provider."), detailRow("Title", normalizeToken(title), "title", "Issue title."), detailRow("Total", itoa(len(cmdCtx.State.Issues)), "count", "Current issue count."))...), nil
	case "open", "close":
		if len(inv.Args) != 2 {
			return Result{}, fmt.Errorf("usage: /issue [status|list|provider <name>|create <title>|open <id>|close <id>|assign <id> <assignee>|label <id> <label>|unlabel <id> <label>]")
		}
		id := strings.TrimSpace(inv.Args[1])
		if id == "" {
			return Result{}, fmt.Errorf("usage: /issue [status|list|provider <name>|create <title>|open <id>|close <id>|assign <id> <assignee>]")
		}
		idx := findIssueIndex(cmdCtx.State.Issues, id)
		if idx < 0 {
			return Result{}, fmt.Errorf("issue not found: %s", id)
		}
		nextStatus := "open"
		if sub == "close" {
			nextStatus = "closed"
		}
		previous := cmdCtx.State.Issues[idx].Status
		cmdCtx.State.Issues[idx].Status = nextStatus
		cmdCtx.State.IssueLastAction = sub
		message := fmt.Sprintf("ISSUE_SET_STATUS\nid=%s\nprevious=%s\nstatus=%s\nchanged=%t", normalizeToken(id), normalizeToken(previous), normalizeToken(nextStatus), !strings.EqualFold(previous, nextStatus))
		return resultWithIntents(message, issueMutationIntents("Issue status", detailRow("ID", normalizeToken(id), normalizeToken(nextStatus), "Issue identifier."), detailRow("Previous", normalizeToken(previous), "previous", "Previous status."), detailRow("Status", normalizeToken(nextStatus), normalizeToken(nextStatus), "Updated status."))...), nil
	case "assign":
		if len(inv.Args) < 3 {
			return Result{}, fmt.Errorf("usage: /issue [status|list|provider <name>|create <title>|open <id>|close <id>|assign <id> <assignee>]")
		}
		id := strings.TrimSpace(inv.Args[1])
		assignee := strings.TrimSpace(strings.Join(inv.Args[2:], " "))
		if id == "" || assignee == "" {
			return Result{}, fmt.Errorf("usage: /issue [status|list|provider <name>|create <title>|open <id>|close <id>|assign <id> <assignee>]")
		}
		idx := findIssueIndex(cmdCtx.State.Issues, id)
		if idx < 0 {
			return Result{}, fmt.Errorf("issue not found: %s", id)
		}
		cmdCtx.State.Issues[idx].Assignee = assignee
		cmdCtx.State.IssueLastAction = "assign"
		message := fmt.Sprintf("ISSUE_ASSIGN\nid=%s\nassignee=%s\nstatus=%s", normalizeToken(id), normalizeToken(assignee), normalizeToken(cmdCtx.State.Issues[idx].Status))
		return resultWithIntents(message, issueMutationIntents("Issue assignee", detailRow("ID", normalizeToken(id), "issue", "Issue identifier."), detailRow("Assignee", normalizeToken(assignee), "assigned", "Assigned owner."), detailRow("Status", normalizeToken(cmdCtx.State.Issues[idx].Status), normalizeToken(cmdCtx.State.Issues[idx].Status), "Current issue status."))...), nil
	case "label", "unlabel":
		if len(inv.Args) < 3 {
			return Result{}, fmt.Errorf("usage: /issue [status|list|provider <name>|create <title>|open <id>|close <id>|assign <id> <assignee>|label <id> <label>|unlabel <id> <label>]")
		}
		id := strings.TrimSpace(inv.Args[1])
		label := strings.TrimSpace(strings.Join(inv.Args[2:], " "))
		if id == "" || label == "" {
			return Result{}, fmt.Errorf("usage: /issue [status|list|provider <name>|create <title>|open <id>|close <id>|assign <id> <assignee>|label <id> <label>|unlabel <id> <label>]")
		}
		idx := findIssueIndex(cmdCtx.State.Issues, id)
		if idx < 0 {
			return Result{}, fmt.Errorf("issue not found: %s", id)
		}
		if sub == "label" {
			before := len(cmdCtx.State.Issues[idx].Labels)
			cmdCtx.State.Issues[idx].Labels = uniqueSortedStrings(append(cmdCtx.State.Issues[idx].Labels, label))
			cmdCtx.State.IssueLastAction = "label"
			message := fmt.Sprintf("ISSUE_LABEL\nid=%s\nlabel=%s\nadded=%t\ncount=%d", normalizeToken(id), normalizeToken(label), len(cmdCtx.State.Issues[idx].Labels) > before, len(cmdCtx.State.Issues[idx].Labels))
			return resultWithIntents(message, issueMutationIntents("Issue label", detailRow("ID", normalizeToken(id), "issue", "Issue identifier."), detailRow("Label", normalizeToken(label), "added", "Label value."), detailRow("Label count", itoa(len(cmdCtx.State.Issues[idx].Labels)), "count", "Total labels on issue."))...), nil
		}
		removed := false
		next := make([]string, 0, len(cmdCtx.State.Issues[idx].Labels))
		for _, existing := range cmdCtx.State.Issues[idx].Labels {
			if strings.EqualFold(strings.TrimSpace(existing), label) {
				removed = true
				continue
			}
			next = append(next, existing)
		}
		cmdCtx.State.Issues[idx].Labels = uniqueSortedStrings(next)
		cmdCtx.State.IssueLastAction = "unlabel"
		message := fmt.Sprintf("ISSUE_UNLABEL\nid=%s\nlabel=%s\nremoved=%t\ncount=%d", normalizeToken(id), normalizeToken(label), removed, len(cmdCtx.State.Issues[idx].Labels))
		return resultWithIntents(message, issueMutationIntents("Issue unlabel", detailRow("ID", normalizeToken(id), "issue", "Issue identifier."), detailRow("Label", normalizeToken(label), boolState(removed, "removed", "missing"), "Label value."), detailRow("Label count", itoa(len(cmdCtx.State.Issues[idx].Labels)), "count", "Remaining labels on issue."))...), nil
	default:
		return Result{}, fmt.Errorf("usage: /issue [status|list|provider <name>|create <title>|open <id>|close <id>|assign <id> <assignee>|label <id> <label>|unlabel <id> <label>]")
	}
}

func executeWorkflowsCommand(cmdCtx Context, inv Invocation) (Result, error) {
	if cmdCtx.State == nil {
		return Result{}, fmt.Errorf("missing command runtime state")
	}
	if len(inv.Args) == 0 || equalFoldTrimmed(inv.Args[0], "status") {
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /workflows [status|list|run <name>|complete <name>|fail <name> [reason]|cancel <name>|rerun <name>]")
		}
		running := 0
		failed := 0
		for _, run := range cmdCtx.State.WorkflowRuns {
			switch strings.ToLower(run.Status) {
			case "running":
				running++
			case "failed":
				failed++
			}
		}
		message := fmt.Sprintf("WORKFLOWS_STATUS\ncount=%d\nrunning=%d\nfailed=%d\nlast_action=%s", len(cmdCtx.State.WorkflowRuns), running, failed, normalizeToken(cmdCtx.State.WorkflowLastAction))
		return resultWithIntents(message, workflowsStatusIntents(cmdCtx.State.WorkflowRuns, cmdCtx.State.WorkflowLastAction)...), nil
	}

	sub := strings.ToLower(strings.TrimSpace(inv.Args[0]))
	switch sub {
	case "list":
		if len(inv.Args) > 1 {
			return Result{}, fmt.Errorf("usage: /workflows [status|list|run <name>|complete <name>|fail <name> [reason]|cancel <name>|rerun <name>]")
		}
		runs := append([]WorkflowRun(nil), cmdCtx.State.WorkflowRuns...)
		sort.Slice(runs, func(i, j int) bool { return runs[i].Name < runs[j].Name })
		lines := []string{"WORKFLOWS_LIST", fmt.Sprintf("count=%d", len(runs))}
		for i, run := range runs {
			idx := i + 1
			lines = append(lines, fmt.Sprintf("workflow.%d.name=%s", idx, normalizeToken(run.Name)))
			lines = append(lines, fmt.Sprintf("workflow.%d.status=%s", idx, normalizeToken(run.Status)))
			lines = append(lines, fmt.Sprintf("workflow.%d.provider=%s", idx, normalizeToken(run.Provider)))
			lines = append(lines, fmt.Sprintf("workflow.%d.last_run=%s", idx, normalizeToken(run.LastRun)))
		}
		return resultWithIntents(strings.Join(lines, "\n"), workflowsListIntents(runs)...), nil
	case "run", "complete", "fail", "cancel", "rerun":
		if len(inv.Args) < 2 {
			return Result{}, fmt.Errorf("usage: /workflows [status|list|run <name>|complete <name>|fail <name> [reason]|cancel <name>|rerun <name>]")
		}
		name := strings.TrimSpace(inv.Args[1])
		if name == "" {
			return Result{}, fmt.Errorf("usage: /workflows [status|list|run <name>|complete <name>|fail <name> [reason]|cancel <name>|rerun <name>]")
		}
		idx := findWorkflowIndex(cmdCtx.State.WorkflowRuns, name)
		if idx < 0 {
			provider := strings.TrimSpace(cmdCtx.State.ProviderName)
			if provider == "" {
				provider = "generic"
			}
			cmdCtx.State.WorkflowRuns = append(cmdCtx.State.WorkflowRuns, WorkflowRun{Name: name, Status: "pending", Provider: provider})
			idx = len(cmdCtx.State.WorkflowRuns) - 1
		}
		reason := "-"
		nextStatus := "running"
		switch sub {
		case "complete":
			nextStatus = "succeeded"
		case "fail":
			nextStatus = "failed"
			if len(inv.Args) > 2 {
				reason = strings.TrimSpace(strings.Join(inv.Args[2:], " "))
				if reason == "" {
					reason = "-"
				}
			}
		case "cancel":
			nextStatus = "cancelled"
		case "rerun":
			nextStatus = "running"
			reason = "rerun"
		}
		cmdCtx.State.WorkflowRuns[idx].Status = nextStatus
		cmdCtx.State.WorkflowRuns[idx].LastRun = time.Now().UTC().Format(time.RFC3339)
		cmdCtx.State.WorkflowLastAction = sub
		message := fmt.Sprintf("WORKFLOWS_SET\nname=%s\nstatus=%s\nreason=%s\nprovider=%s", normalizeToken(name), normalizeToken(nextStatus), normalizeToken(reason), normalizeToken(cmdCtx.State.WorkflowRuns[idx].Provider))
		return resultWithIntents(message, workflowsMutationIntents(name, nextStatus, reason, cmdCtx.State.WorkflowRuns[idx].Provider)...), nil
	default:
		return Result{}, fmt.Errorf("usage: /workflows [status|list|run <name>|complete <name>|fail <name> [reason]|cancel <name>|rerun <name>]")
	}
}
