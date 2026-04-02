package state

import "strings"

type SessionMetadataPageQuery struct {
	SessionMetadataQuery
	Offset         int
	Limit          int
	ProjectAny     []string
	MinMessages    int
	MinEvents      int
	LastRole       string
	HasRuntime     *bool
	RequireSummary bool
}

type SessionMetadataPage struct {
	Items   []SessionMetadataSnapshot `json:"items,omitempty"`
	Total   int                       `json:"total"`
	Offset  int                       `json:"offset"`
	Limit   int                       `json:"limit"`
	HasMore bool                      `json:"has_more"`
}

type SessionMetadataSummary struct {
	Total               int            `json:"total"`
	TotalMessages       int            `json:"total_messages"`
	TotalEvents         int            `json:"total_events"`
	DistinctProjects    int            `json:"distinct_projects"`
	SessionsWithSummary int            `json:"sessions_with_summary"`
	RoleCounts          map[string]int `json:"role_counts,omitempty"`
	RuntimeCounts       map[string]int `json:"runtime_counts,omitempty"`
}

func (r *SessionMetadataRepository) QueryPage(query SessionMetadataPageQuery) (SessionMetadataPage, error) {
	items, err := r.Query(query.SessionMetadataQuery)
	if err != nil {
		return SessionMetadataPage{}, err
	}
	items = filterSessionMetadataSnapshots(items, query)
	total := len(items)
	start, end := paginateBounds(total, query.Offset, query.Limit)
	return SessionMetadataPage{Items: items[start:end], Total: total, Offset: start, Limit: end - start, HasMore: end < total}, nil
}

func (r *SessionMetadataRepository) QuerySummary(query SessionMetadataPageQuery) (SessionMetadataSummary, error) {
	items, err := r.Query(query.SessionMetadataQuery)
	if err != nil {
		return SessionMetadataSummary{}, err
	}
	return SummarizeSessionMetadataSnapshots(filterSessionMetadataSnapshots(items, query)), nil
}

func SummarizeSessionMetadataSnapshots(items []SessionMetadataSnapshot) SessionMetadataSummary {
	out := SessionMetadataSummary{
		RoleCounts:    make(map[string]int),
		RuntimeCounts: make(map[string]int),
	}
	projects := make(map[string]struct{})
	for _, item := range items {
		out.Total++
		out.TotalMessages += item.Metadata.MessageCount
		out.TotalEvents += item.Metadata.EventCount
		if project := strings.TrimSpace(item.Metadata.ProjectPath); project != "" {
			projects[project] = struct{}{}
		}
		if strings.TrimSpace(item.Metadata.Summary) != "" {
			out.SessionsWithSummary++
		}
		if role := strings.TrimSpace(item.Metadata.LastRole); role != "" {
			out.RoleCounts[role]++
		}
		if runtime := strings.TrimSpace(item.Metadata.RuntimeSurface); runtime != "" {
			out.RuntimeCounts[runtime]++
		}
	}
	out.DistinctProjects = len(projects)
	if len(out.RoleCounts) == 0 {
		out.RoleCounts = nil
	}
	if len(out.RuntimeCounts) == 0 {
		out.RuntimeCounts = nil
	}
	return out
}

func filterSessionMetadataSnapshots(items []SessionMetadataSnapshot, query SessionMetadataPageQuery) []SessionMetadataSnapshot {
	if len(items) == 0 {
		return nil
	}
	role := normalizeEquals(query.LastRole)
	projects := normalizeSliceContains(query.ProjectAny)
	out := make([]SessionMetadataSnapshot, 0, len(items))
	for _, item := range items {
		if len(projects) > 0 && !matchAnyPath(item.Metadata.ProjectPath, projects) {
			continue
		}
		if query.MinMessages > 0 && item.Metadata.MessageCount < query.MinMessages {
			continue
		}
		if query.MinEvents > 0 && item.Metadata.EventCount < query.MinEvents {
			continue
		}
		hasRuntime := normalizeContains(item.Metadata.RuntimeSurface) != ""
		if query.HasRuntime != nil && hasRuntime != *query.HasRuntime {
			continue
		}
		if role != "" && normalizeEquals(item.Metadata.LastRole) != role {
			continue
		}
		if query.RequireSummary && normalizeContains(item.Metadata.Summary) == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}
