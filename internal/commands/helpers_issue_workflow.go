package commands

import "strings"

func findIssueIndex(issues []IssueRecord, id string) int {
	for i, issue := range issues {
		if strings.EqualFold(strings.TrimSpace(issue.ID), strings.TrimSpace(id)) {
			return i
		}
	}
	return -1
}

func findWorkflowIndex(runs []WorkflowRun, name string) int {
	for i, run := range runs {
		if strings.EqualFold(strings.TrimSpace(run.Name), strings.TrimSpace(name)) {
			return i
		}
	}
	return -1
}

func findShareLinkIndex(links []ShareLink, id string) int {
	for i, link := range links {
		if strings.EqualFold(strings.TrimSpace(link.ID), strings.TrimSpace(id)) {
			return i
		}
	}
	return -1
}
