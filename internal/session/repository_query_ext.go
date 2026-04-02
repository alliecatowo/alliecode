package session

import "strings"

type PageQuery struct {
	Query
	Offset            int
	Limit             int
	StartedAfterUnix  int64
	StartedBeforeUnix int64
	MaxMessages       int
	MinEvents         int
	MaxEvents         int
	LastRoleContains  string
	RequireRuntime    bool
	RequireProvider   bool
	RequireModel      bool
}

type Page struct {
	Items   []Transcript `json:"items,omitempty"`
	Total   int          `json:"total"`
	Offset  int          `json:"offset"`
	Limit   int          `json:"limit"`
	HasMore bool         `json:"has_more"`
}

type Summary struct {
	Total            int            `json:"total"`
	TotalMessages    int            `json:"total_messages"`
	TotalEvents      int            `json:"total_events"`
	ProviderCount    map[string]int `json:"provider_count,omitempty"`
	RuntimeCount     map[string]int `json:"runtime_count,omitempty"`
	DistinctProjects int            `json:"distinct_projects"`
	DistinctModels   int            `json:"distinct_models"`
	DistinctSessions int            `json:"distinct_sessions"`
	FirstStartedUnix int64          `json:"first_started_unix,omitempty"`
	LastStartedUnix  int64          `json:"last_started_unix,omitempty"`
}

func (r *Repository) QueryPage(query PageQuery) (Page, error) {
	base := query.Query
	base.Limit = 0
	items, err := r.Query(base)
	if err != nil {
		return Page{}, err
	}
	items = filterTranscripts(items, query)
	total := len(items)
	start, end := pageBounds(total, query.Offset, query.Limit)
	return Page{Items: items[start:end], Total: total, Offset: start, Limit: end - start, HasMore: end < total}, nil
}

func (r *Repository) QuerySummary(query PageQuery) (Summary, error) {
	base := query.Query
	base.Limit = 0
	items, err := r.Query(base)
	if err != nil {
		return Summary{}, err
	}
	return summarizeTranscripts(filterTranscripts(items, query)), nil
}

func filterTranscripts(items []Transcript, query PageQuery) []Transcript {
	if len(items) == 0 {
		return nil
	}
	lastRoleContains := strings.ToLower(strings.TrimSpace(query.LastRoleContains))
	out := make([]Transcript, 0, len(items))
	for _, item := range items {
		if query.StartedAfterUnix > 0 && item.FirstEventAt.Unix() < query.StartedAfterUnix {
			continue
		}
		if query.StartedBeforeUnix > 0 && item.FirstEventAt.Unix() > query.StartedBeforeUnix {
			continue
		}
		if query.MaxMessages > 0 && item.MessageCount > query.MaxMessages {
			continue
		}
		if query.MinEvents > 0 && item.EventCount < query.MinEvents {
			continue
		}
		if query.MaxEvents > 0 && item.EventCount > query.MaxEvents {
			continue
		}
		if lastRoleContains != "" && !strings.Contains(strings.ToLower(strings.TrimSpace(item.LastMessageRole)), lastRoleContains) {
			continue
		}
		if query.RequireRuntime && strings.TrimSpace(item.RuntimeSurface) == "" {
			continue
		}
		if query.RequireProvider && strings.TrimSpace(item.Provider) == "" {
			continue
		}
		if query.RequireModel && strings.TrimSpace(item.Model) == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

func summarizeTranscripts(items []Transcript) Summary {
	out := Summary{ProviderCount: make(map[string]int), RuntimeCount: make(map[string]int)}
	projects := make(map[string]struct{})
	models := make(map[string]struct{})
	sessions := make(map[string]struct{})
	for _, item := range items {
		out.Total++
		out.TotalMessages += item.MessageCount
		out.TotalEvents += item.EventCount
		started := item.FirstEventAt.Unix()
		if out.FirstStartedUnix == 0 || started < out.FirstStartedUnix {
			out.FirstStartedUnix = started
		}
		if started > out.LastStartedUnix {
			out.LastStartedUnix = started
		}
		if v := strings.TrimSpace(item.Provider); v != "" {
			out.ProviderCount[v]++
		}
		if v := strings.TrimSpace(item.RuntimeSurface); v != "" {
			out.RuntimeCount[v]++
		}
		if v := strings.TrimSpace(item.ProjectPath); v != "" {
			projects[v] = struct{}{}
		}
		if v := strings.TrimSpace(item.Model); v != "" {
			models[v] = struct{}{}
		}
		if v := strings.TrimSpace(item.SessionID); v != "" {
			sessions[v] = struct{}{}
		}
	}
	out.DistinctProjects = len(projects)
	out.DistinctModels = len(models)
	out.DistinctSessions = len(sessions)
	if len(out.ProviderCount) == 0 {
		out.ProviderCount = nil
	}
	if len(out.RuntimeCount) == 0 {
		out.RuntimeCount = nil
	}
	return out
}

func pageBounds(total, offset, limit int) (int, int) {
	if total <= 0 {
		return 0, 0
	}
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	end := total
	if limit > 0 && offset+limit < end {
		end = offset + limit
	}
	return offset, end
}
