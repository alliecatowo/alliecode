package history

type IndexPage struct {
	Items   []IndexRecord `json:"items,omitempty"`
	Total   int           `json:"total"`
	Offset  int           `json:"offset"`
	Limit   int           `json:"limit"`
	HasMore bool          `json:"has_more"`
}

type IndexSummary struct {
	Total              int            `json:"total"`
	TypeCounts         map[string]int `json:"type_counts,omitempty"`
	Projects           map[string]int `json:"projects,omitempty"`
	ByRuntime          map[string]int `json:"by_runtime,omitempty"`
	ByCommand          map[string]int `json:"by_command,omitempty"`
	DistinctSessions   int            `json:"distinct_sessions"`
	FirstTimestampUnix int64          `json:"first_timestamp_unix,omitempty"`
	LastTimestampUnix  int64          `json:"last_timestamp_unix,omitempty"`
}

func (r *Repository) BuildIndexPage(query HistoryQuery, indexQuery IndexQuery, offset, limit int) (IndexPage, error) {
	indexQuery.Limit = 0
	items, err := r.BuildIndex(query, indexQuery)
	if err != nil {
		return IndexPage{}, err
	}
	total := len(items)
	start, end := pagination(total, offset, limit)
	return IndexPage{Items: items[start:end], Total: total, Offset: start, Limit: end - start, HasMore: end < total}, nil
}

func SummarizeIndex(items []IndexRecord) IndexSummary {
	out := IndexSummary{TypeCounts: make(map[string]int), Projects: make(map[string]int), ByRuntime: make(map[string]int), ByCommand: make(map[string]int)}
	sessions := make(map[string]struct{})
	for _, item := range items {
		out.Total++
		out.TypeCounts[string(item.Type)]++
		if item.Project != "" {
			out.Projects[item.Project]++
		}
		if item.RuntimeSurface != "" {
			out.ByRuntime[item.RuntimeSurface]++
		}
		if item.CommandSurface != "" {
			out.ByCommand[item.CommandSurface]++
		}
		if item.SessionID != "" {
			sessions[item.SessionID] = struct{}{}
		}
		ts := item.Timestamp.Unix()
		if out.FirstTimestampUnix == 0 || ts < out.FirstTimestampUnix {
			out.FirstTimestampUnix = ts
		}
		if ts > out.LastTimestampUnix {
			out.LastTimestampUnix = ts
		}
	}
	out.DistinctSessions = len(sessions)
	if len(out.TypeCounts) == 0 {
		out.TypeCounts = nil
	}
	if len(out.Projects) == 0 {
		out.Projects = nil
	}
	if len(out.ByRuntime) == 0 {
		out.ByRuntime = nil
	}
	if len(out.ByCommand) == 0 {
		out.ByCommand = nil
	}
	return out
}
