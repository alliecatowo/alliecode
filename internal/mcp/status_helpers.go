package mcp

import "sort"

type StatusSummary struct {
	Total            int            `json:"total"`
	ByConnection     map[string]int `json:"by_connection"`
	ByAuthStatus     map[string]int `json:"by_auth_status"`
	Authenticated    int            `json:"authenticated"`
	NeedsAuth        int            `json:"needs_auth"`
	Failed           int            `json:"failed"`
	Disabled         int            `json:"disabled"`
	WithRecentError  int            `json:"with_recent_error"`
	WithRecentAccess int            `json:"with_recent_success"`
}

func sortServerStatuses(statuses []ServerStatus) {
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].ServerName < statuses[j].ServerName
	})
}

func cloneServerStatus(status ServerStatus) ServerStatus {
	normalizeServerStatusFields(&status)
	return status
}

func summarizeServerStatuses(statuses []ServerStatus) StatusSummary {
	out := StatusSummary{
		ByConnection: map[string]int{},
		ByAuthStatus: map[string]int{},
	}
	for _, status := range statuses {
		out.Total++
		out.ByConnection[string(status.ConnectionState)]++
		out.ByAuthStatus[string(status.AuthStatus)]++
		if status.Authenticated {
			out.Authenticated++
		}
		if status.ConnectionState == ServerConnectionNeedsAuth {
			out.NeedsAuth++
		}
		if status.ConnectionState == ServerConnectionFailed {
			out.Failed++
		}
		if status.ConnectionState == ServerConnectionDisabled {
			out.Disabled++
		}
		if !status.LastErrorAt.IsZero() {
			out.WithRecentError++
		}
		if !status.LastSuccessAt.IsZero() {
			out.WithRecentAccess++
		}
	}
	return out
}
