package tasks

import "strings"

type TaskQuery struct {
	Statuses        []Status `json:"statuses,omitempty"`
	Owner           string   `json:"owner,omitempty"`
	Limit           int      `json:"limit,omitempty"`
	IncludeTerminal *bool    `json:"include_terminal,omitempty"`
}

type TaskQuerySummary struct {
	Query           TaskQuery      `json:"query"`
	Scanned         int            `json:"scanned"`
	Matched         int            `json:"matched"`
	Returned        int            `json:"returned"`
	FilteredOut     int            `json:"filtered_out"`
	Truncated       bool           `json:"truncated"`
	ByStatus        map[Status]int `json:"by_status,omitempty"`
	ByOwner         map[string]int `json:"by_owner,omitempty"`
	OwnerFilter     string         `json:"owner_filter,omitempty"`
	HasStatusFilter bool           `json:"has_status_filter"`
	IncludeTerminal bool           `json:"include_terminal"`
	RequestedLimit  int            `json:"requested_limit,omitempty"`
}

func ApplyTaskQuery(items []Task, query TaskQuery) ([]Task, TaskQuerySummary) {
	out := make([]Task, 0, len(items))
	statusFilter := map[Status]bool{}
	for _, status := range query.Statuses {
		statusFilter[status] = true
	}
	ownerFilter := strings.ToLower(strings.TrimSpace(query.Owner))
	includeTerminal := true
	if query.IncludeTerminal != nil {
		includeTerminal = *query.IncludeTerminal
	}

	for _, item := range items {
		summaryScanned := true
		if len(statusFilter) > 0 && !statusFilter[item.Status] {
			if summaryScanned {
				summaryScanned = false
			}
			continue
		}
		if ownerFilter != "" && strings.ToLower(strings.TrimSpace(item.Owner)) != ownerFilter {
			if summaryScanned {
				summaryScanned = false
			}
			continue
		}
		if !includeTerminal && isTerminalStatus(item.Status) {
			if summaryScanned {
				summaryScanned = false
			}
			continue
		}
		out = append(out, item)
	}

	summary := TaskQuerySummary{
		Query:           query,
		Scanned:         len(items),
		Matched:         len(out),
		Returned:        len(out),
		FilteredOut:     len(items) - len(out),
		ByStatus:        map[Status]int{},
		ByOwner:         map[string]int{},
		OwnerFilter:     ownerFilter,
		HasStatusFilter: len(statusFilter) > 0,
		IncludeTerminal: includeTerminal,
		RequestedLimit:  query.Limit,
	}
	for _, item := range out {
		summary.ByStatus[item.Status]++
		if strings.TrimSpace(item.Owner) != "" {
			summary.ByOwner[item.Owner]++
		}
	}

	if query.Limit > 0 && len(out) > query.Limit {
		out = out[:query.Limit]
		summary.Returned = len(out)
		summary.Truncated = true
	}
	if len(summary.ByStatus) == 0 {
		summary.ByStatus = nil
	}
	if len(summary.ByOwner) == 0 {
		summary.ByOwner = nil
	}

	return out, summary
}
