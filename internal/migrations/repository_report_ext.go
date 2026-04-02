package migrations

import "time"

type QueryPage struct {
	Items   []LedgerEntry `json:"items,omitempty"`
	Total   int           `json:"total"`
	Offset  int           `json:"offset"`
	Limit   int           `json:"limit"`
	HasMore bool          `json:"has_more"`
}

type Report struct {
	Summary
	FirstAppliedAt  time.Time        `json:"first_applied_at,omitempty"`
	LastAppliedAt   time.Time        `json:"last_applied_at,omitempty"`
	LastVersion     string           `json:"last_version,omitempty"`
	AppliedVersions []string         `json:"applied_versions,omitempty"`
	Statuses        map[string]int   `json:"statuses,omitempty"`
	Durations       map[string]int64 `json:"durations,omitempty"`
}

func (r *Repository) QueryPage(query Query, offset int) (QueryPage, error) {
	limit := query.Limit
	query.Limit = 0
	items, err := r.Query(query)
	if err != nil {
		return QueryPage{}, err
	}
	total := len(items)
	start, end := pagination(total, offset, limit)
	return QueryPage{Items: items[start:end], Total: total, Offset: start, Limit: end - start, HasMore: end < total}, nil
}

func (r *Repository) QueryReport(query Query) (Report, error) {
	query.Limit = 0
	items, err := r.Query(query)
	if err != nil {
		return Report{}, err
	}
	return BuildReport(items), nil
}

func BuildReport(entries []LedgerEntry) Report {
	summary := SummarizeEntries(entries)
	out := Report{Summary: summary, Statuses: make(map[string]int), Durations: map[string]int64{"min_ms": 0, "max_ms": 0, "avg_ms": 0}}
	if len(entries) == 0 {
		out.Statuses = nil
		return out
	}
	var minDuration int64
	for i, entry := range entries {
		if i == 0 || entry.AppliedAt.Before(out.FirstAppliedAt) {
			out.FirstAppliedAt = entry.AppliedAt
		}
		if entry.AppliedAt.After(out.LastAppliedAt) {
			out.LastAppliedAt = entry.AppliedAt
			out.LastVersion = entry.Version
		}
		out.Statuses[entry.Status]++
		if entry.Status == "applied" {
			out.AppliedVersions = append(out.AppliedVersions, entry.Version)
		}
		if i == 0 || entry.DurationMs < minDuration {
			minDuration = entry.DurationMs
		}
		if entry.DurationMs > out.Durations["max_ms"] {
			out.Durations["max_ms"] = entry.DurationMs
		}
	}
	out.Durations["min_ms"] = minDuration
	out.Durations["avg_ms"] = summary.TotalDurationMs / int64(len(entries))
	return out
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
