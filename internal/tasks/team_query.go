package tasks

import "strings"

type TeamQuery struct {
	Statuses []TeamStatus `json:"statuses,omitempty"`
	Member   string       `json:"member,omitempty"`
	Limit    int          `json:"limit,omitempty"`
}

type TeamQuerySummary struct {
	Query        TeamQuery             `json:"query"`
	Scanned      int                   `json:"scanned"`
	Matched      int                   `json:"matched"`
	Returned     int                   `json:"returned"`
	FilteredOut  int                   `json:"filtered_out"`
	Truncated    bool                  `json:"truncated"`
	ByStatus     map[TeamStatus]int    `json:"by_status,omitempty"`
	ByMember     map[string]int        `json:"by_member,omitempty"`
	ByTeamName   map[string]TeamStatus `json:"by_team_name,omitempty"`
	MemberFilter string                `json:"member_filter,omitempty"`
}

func ApplyTeamQuery(teams []Team, query TeamQuery) ([]Team, TeamQuerySummary) {
	statusFilter := map[TeamStatus]bool{}
	for _, status := range query.Statuses {
		statusFilter[status] = true
	}
	memberFilter := strings.ToLower(strings.TrimSpace(query.Member))

	filtered := make([]Team, 0, len(teams))
	for _, team := range teams {
		if len(statusFilter) > 0 && !statusFilter[team.Status] {
			continue
		}
		if memberFilter != "" {
			ok := false
			for _, member := range team.Members {
				if strings.ToLower(strings.TrimSpace(member.Name)) == memberFilter {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
		}
		filtered = append(filtered, team)
	}

	summary := TeamQuerySummary{
		Query:        query,
		Scanned:      len(teams),
		Matched:      len(filtered),
		Returned:     len(filtered),
		FilteredOut:  len(teams) - len(filtered),
		ByStatus:     map[TeamStatus]int{},
		ByMember:     map[string]int{},
		ByTeamName:   map[string]TeamStatus{},
		MemberFilter: memberFilter,
	}
	for _, team := range filtered {
		summary.ByStatus[team.Status]++
		summary.ByTeamName[team.Name] = team.Status
		for _, member := range team.Members {
			if strings.TrimSpace(member.Name) != "" {
				summary.ByMember[member.Name]++
			}
		}
	}

	if query.Limit > 0 && len(filtered) > query.Limit {
		filtered = filtered[:query.Limit]
		summary.Returned = len(filtered)
		summary.Truncated = true
	}
	if len(summary.ByStatus) == 0 {
		summary.ByStatus = nil
	}
	if len(summary.ByMember) == 0 {
		summary.ByMember = nil
	}
	if len(summary.ByTeamName) == 0 {
		summary.ByTeamName = nil
	}

	return filtered, summary
}
