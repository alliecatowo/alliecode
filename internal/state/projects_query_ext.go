package state

import "strings"

type ProjectPageQuery struct {
	ProjectQuery
	Offset           int
	Limit            int
	PathContainsAny  []string
	MinOpenCount     int
	MinDistinctCount int
	OpenedAfterUnix  int64
	OpenedBeforeUnix int64
	RequireSession   bool
}

type ProjectPage struct {
	Items   []ProjectSnapshot `json:"items,omitempty"`
	Total   int               `json:"total"`
	Offset  int               `json:"offset"`
	Limit   int               `json:"limit"`
	HasMore bool              `json:"has_more"`
}

type ProjectSummary struct {
	Total                int            `json:"total"`
	TotalOpens           int            `json:"total_opens"`
	DistinctSessionTotal int            `json:"distinct_session_total"`
	MostRecentOpenedAt   int64          `json:"most_recent_opened_at,omitempty"`
	OldestOpenedAt       int64          `json:"oldest_opened_at,omitempty"`
	ProjectsWithSession  int            `json:"projects_with_session"`
	ProviderCounts       map[string]int `json:"provider_counts,omitempty"`
	RuntimeCounts        map[string]int `json:"runtime_counts,omitempty"`
	CommandCounts        map[string]int `json:"command_counts,omitempty"`
}

func (r *ProjectsRepository) QueryPage(query ProjectPageQuery) (ProjectPage, error) {
	items, err := r.Query(query.ProjectQuery)
	if err != nil {
		return ProjectPage{}, err
	}
	items = filterProjectSnapshots(items, query)
	total := len(items)
	start, end := paginateBounds(total, query.Offset, query.Limit)
	page := ProjectPage{
		Items:   items[start:end],
		Total:   total,
		Offset:  start,
		Limit:   end - start,
		HasMore: end < total,
	}
	return page, nil
}

func (r *ProjectsRepository) QuerySummary(query ProjectPageQuery) (ProjectSummary, error) {
	items, err := r.Query(query.ProjectQuery)
	if err != nil {
		return ProjectSummary{}, err
	}
	return SummarizeProjectSnapshots(filterProjectSnapshots(items, query)), nil
}

func SummarizeProjectSnapshots(items []ProjectSnapshot) ProjectSummary {
	out := ProjectSummary{
		ProviderCounts: make(map[string]int),
		RuntimeCounts:  make(map[string]int),
		CommandCounts:  make(map[string]int),
	}
	for _, item := range items {
		out.Total++
		out.TotalOpens += item.Metadata.OpenCount
		out.DistinctSessionTotal += item.Metadata.DistinctSessionCount
		if item.Metadata.LastOpenedAt > out.MostRecentOpenedAt {
			out.MostRecentOpenedAt = item.Metadata.LastOpenedAt
		}
		if out.OldestOpenedAt == 0 || item.Metadata.LastOpenedAt < out.OldestOpenedAt {
			out.OldestOpenedAt = item.Metadata.LastOpenedAt
		}
		if strings.TrimSpace(item.Metadata.LastSessionID) != "" {
			out.ProjectsWithSession++
		}
		if v := strings.TrimSpace(item.Metadata.LastProvider); v != "" {
			out.ProviderCounts[v]++
		}
		if v := strings.TrimSpace(item.Metadata.LastRuntimeSurface); v != "" {
			out.RuntimeCounts[v]++
		}
		if v := strings.TrimSpace(item.Metadata.LastCommandSurface); v != "" {
			out.CommandCounts[v]++
		}
	}
	if len(out.ProviderCounts) == 0 {
		out.ProviderCounts = nil
	}
	if len(out.RuntimeCounts) == 0 {
		out.RuntimeCounts = nil
	}
	if len(out.CommandCounts) == 0 {
		out.CommandCounts = nil
	}
	return out
}

func filterProjectSnapshots(items []ProjectSnapshot, query ProjectPageQuery) []ProjectSnapshot {
	if len(items) == 0 {
		return nil
	}
	pathContainsAny := normalizeSliceContains(query.PathContainsAny)
	out := make([]ProjectSnapshot, 0, len(items))
	for _, item := range items {
		if len(pathContainsAny) > 0 && !matchAnyPath(item.Path, pathContainsAny) {
			continue
		}
		if query.MinOpenCount > 0 && item.Metadata.OpenCount < query.MinOpenCount {
			continue
		}
		if query.MinDistinctCount > 0 && item.Metadata.DistinctSessionCount < query.MinDistinctCount {
			continue
		}
		if query.OpenedAfterUnix > 0 && item.Metadata.LastOpenedAt < query.OpenedAfterUnix {
			continue
		}
		if query.OpenedBeforeUnix > 0 && item.Metadata.LastOpenedAt > query.OpenedBeforeUnix {
			continue
		}
		if query.RequireSession && strings.TrimSpace(item.Metadata.LastSessionID) == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func normalizeSliceContains(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := normalizeContains(value)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func matchAnyPath(path string, contains []string) bool {
	path = normalizeContains(path)
	for _, part := range contains {
		if strings.Contains(path, part) {
			return true
		}
	}
	return false
}
