package history

import "strings"

type EventPageQuery struct {
	HistoryQuery
	Offset          int
	Limit           int
	ProjectContains string
	SessionContains string
	Provider        string
	Model           string
	RuntimeSurface  string
	CommandSurface  string
	DisplayContains string
	DisplayAny      []string
	TagContains     string
	HasTags         *bool
	MinTagCount     int
	RequireModel    bool
}

type EventPage struct {
	Items   []Event `json:"items,omitempty"`
	Total   int     `json:"total"`
	Offset  int     `json:"offset"`
	Limit   int     `json:"limit"`
	HasMore bool    `json:"has_more"`
}

type EventSummary struct {
	Total              int            `json:"total"`
	ByType             map[string]int `json:"by_type,omitempty"`
	ByProvider         map[string]int `json:"by_provider,omitempty"`
	ByRuntime          map[string]int `json:"by_runtime,omitempty"`
	ByCommand          map[string]int `json:"by_command,omitempty"`
	ByModel            map[string]int `json:"by_model,omitempty"`
	DistinctTags       int            `json:"distinct_tags"`
	DistinctUsers      int            `json:"distinct_sessions"`
	DistinctProjects   int            `json:"distinct_projects"`
	FirstTimestampUnix int64          `json:"first_timestamp_unix,omitempty"`
	LastTimestampUnix  int64          `json:"last_timestamp_unix,omitempty"`
}

func (r *Repository) QueryEventsPage(query EventPageQuery) (EventPage, error) {
	base := query.HistoryQuery
	base.MaxItems = 0
	items, err := r.QueryEvents(base)
	if err != nil {
		return EventPage{}, err
	}
	items = filterEvents(items, query)
	total := len(items)
	start, end := pagination(total, query.Offset, query.Limit)
	return EventPage{Items: items[start:end], Total: total, Offset: start, Limit: end - start, HasMore: end < total}, nil
}

func (r *Repository) QueryEventsSummary(query EventPageQuery) (EventSummary, error) {
	base := query.HistoryQuery
	base.MaxItems = 0
	items, err := r.QueryEvents(base)
	if err != nil {
		return EventSummary{}, err
	}
	return summarizeEvents(filterEvents(items, query)), nil
}

func filterEvents(items []Event, query EventPageQuery) []Event {
	if len(items) == 0 {
		return nil
	}
	sessionContains := strings.ToLower(strings.TrimSpace(query.SessionContains))
	provider := strings.ToLower(strings.TrimSpace(query.Provider))
	model := strings.ToLower(strings.TrimSpace(query.Model))
	runtime := strings.ToLower(strings.TrimSpace(query.RuntimeSurface))
	command := strings.ToLower(strings.TrimSpace(query.CommandSurface))
	projectContains := strings.ToLower(strings.TrimSpace(query.ProjectContains))
	displayContains := strings.ToLower(strings.TrimSpace(query.DisplayContains))
	displayAny := normalizeContainsAny(query.DisplayAny)
	tagContains := strings.ToLower(strings.TrimSpace(query.TagContains))

	out := make([]Event, 0, len(items))
	for _, item := range items {
		if projectContains != "" && !strings.Contains(strings.ToLower(strings.TrimSpace(item.Project)), projectContains) {
			continue
		}
		if sessionContains != "" && !strings.Contains(strings.ToLower(strings.TrimSpace(item.SessionID)), sessionContains) {
			continue
		}
		if provider != "" && strings.ToLower(strings.TrimSpace(item.Provider)) != provider {
			continue
		}
		if model != "" && strings.ToLower(strings.TrimSpace(item.Model)) != model {
			continue
		}
		if runtime != "" && strings.ToLower(strings.TrimSpace(item.RuntimeSurface)) != runtime {
			continue
		}
		if command != "" && strings.ToLower(strings.TrimSpace(item.CommandSurface)) != command {
			continue
		}
		if query.RequireModel && strings.TrimSpace(item.Model) == "" {
			continue
		}
		display := strings.ToLower(strings.TrimSpace(item.Display))
		if displayContains != "" && !strings.Contains(display, displayContains) {
			continue
		}
		if len(displayAny) > 0 && !containsAny(display, displayAny) {
			continue
		}
		hasTags := len(item.Tags) > 0
		if query.HasTags != nil && hasTags != *query.HasTags {
			continue
		}
		if query.MinTagCount > 0 && len(item.Tags) < query.MinTagCount {
			continue
		}
		if tagContains != "" && !eventHasTag(item.Tags, tagContains) {
			continue
		}
		out = append(out, item)
	}
	return out
}

func summarizeEvents(items []Event) EventSummary {
	out := EventSummary{ByType: make(map[string]int), ByProvider: make(map[string]int), ByRuntime: make(map[string]int), ByCommand: make(map[string]int), ByModel: make(map[string]int)}
	tagSet := map[string]struct{}{}
	sessionSet := map[string]struct{}{}
	projectSet := map[string]struct{}{}
	for _, item := range items {
		out.Total++
		out.ByType[string(item.Type)]++
		ts := item.Timestamp.Unix()
		if out.FirstTimestampUnix == 0 || ts < out.FirstTimestampUnix {
			out.FirstTimestampUnix = ts
		}
		if ts > out.LastTimestampUnix {
			out.LastTimestampUnix = ts
		}
		if v := strings.TrimSpace(item.Provider); v != "" {
			out.ByProvider[v]++
		}
		if v := strings.TrimSpace(item.RuntimeSurface); v != "" {
			out.ByRuntime[v]++
		}
		if v := strings.TrimSpace(item.CommandSurface); v != "" {
			out.ByCommand[v]++
		}
		if v := strings.TrimSpace(item.Model); v != "" {
			out.ByModel[v]++
		}
		for _, tag := range item.Tags {
			trimmed := strings.TrimSpace(tag)
			if trimmed != "" {
				tagSet[trimmed] = struct{}{}
			}
		}
		if id := strings.TrimSpace(item.SessionID); id != "" {
			sessionSet[id] = struct{}{}
		}
		if project := strings.TrimSpace(item.Project); project != "" {
			projectSet[project] = struct{}{}
		}
	}
	out.DistinctTags = len(tagSet)
	out.DistinctUsers = len(sessionSet)
	out.DistinctProjects = len(projectSet)
	if len(out.ByType) == 0 {
		out.ByType = nil
	}
	if len(out.ByProvider) == 0 {
		out.ByProvider = nil
	}
	if len(out.ByRuntime) == 0 {
		out.ByRuntime = nil
	}
	if len(out.ByCommand) == 0 {
		out.ByCommand = nil
	}
	if len(out.ByModel) == 0 {
		out.ByModel = nil
	}
	return out
}

func normalizeContainsAny(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.ToLower(strings.TrimSpace(value))
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func containsAny(value string, parts []string) bool {
	for _, part := range parts {
		if strings.Contains(value, part) {
			return true
		}
	}
	return false
}

func pagination(total, offset, limit int) (int, int) {
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
