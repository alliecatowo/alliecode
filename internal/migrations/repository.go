package migrations

import (
	"sort"
	"strings"
	"time"
)

type Repository struct {
	ledger *Ledger
}

type Query struct {
	VersionPrefix       string
	VersionSuffix       string
	Description         string
	DescriptionContains string
	Status              string
	StatusIn            []string
	MinDurationMs       int64
	MaxDurationMs       int64
	AppliedAfter        time.Time
	AppliedBefore       time.Time
	Limit               int
}

type Summary struct {
	Total           int   `json:"total"`
	Applied         int   `json:"applied"`
	Skipped         int   `json:"skipped"`
	Failed          int   `json:"failed"`
	TotalDurationMs int64 `json:"total_duration_ms"`
}

func NewRepository(ledger *Ledger) *Repository {
	return &Repository{ledger: ledger}
}

func (r *Repository) Ledger() *Ledger {
	if r == nil {
		return nil
	}
	return r.ledger
}

func (r *Repository) Entries() ([]LedgerEntry, error) {
	if r == nil || r.ledger == nil {
		return nil, nil
	}
	return r.ledger.Entries()
}

func (r *Repository) Query(query Query) ([]LedgerEntry, error) {
	if r == nil || r.ledger == nil {
		return nil, nil
	}
	entries, err := r.ledger.Entries()
	if err != nil {
		return nil, err
	}
	return FilterEntries(entries, query), nil
}

func FilterEntries(entries []LedgerEntry, query Query) []LedgerEntry {
	if len(entries) == 0 {
		return nil
	}
	versionPrefix := strings.TrimSpace(query.VersionPrefix)
	versionSuffix := strings.TrimSpace(query.VersionSuffix)
	description := strings.ToLower(strings.TrimSpace(query.Description))
	descriptionContains := strings.ToLower(strings.TrimSpace(query.DescriptionContains))
	status := strings.ToLower(strings.TrimSpace(query.Status))
	statuses := normalizeStatuses(query.StatusIn)

	out := make([]LedgerEntry, 0, len(entries))
	for _, entry := range entries {
		if versionPrefix != "" && !strings.HasPrefix(entry.Version, versionPrefix) {
			continue
		}
		if versionSuffix != "" && !strings.HasSuffix(entry.Version, versionSuffix) {
			continue
		}
		if description != "" && !strings.Contains(strings.ToLower(entry.Description), description) {
			continue
		}
		if descriptionContains != "" && !strings.Contains(strings.ToLower(entry.Description), descriptionContains) {
			continue
		}
		if status != "" && strings.ToLower(strings.TrimSpace(entry.Status)) != status {
			continue
		}
		if len(statuses) > 0 && !statusAllowed(entry.Status, statuses) {
			continue
		}
		if query.MinDurationMs > 0 && entry.DurationMs < query.MinDurationMs {
			continue
		}
		if query.MaxDurationMs > 0 && entry.DurationMs > query.MaxDurationMs {
			continue
		}
		if !query.AppliedAfter.IsZero() && entry.AppliedAt.Before(query.AppliedAfter) {
			continue
		}
		if !query.AppliedBefore.IsZero() && entry.AppliedAt.After(query.AppliedBefore) {
			continue
		}
		out = append(out, entry)
	}

	sort.SliceStable(out, func(i, j int) bool {
		return out[i].AppliedAt.After(out[j].AppliedAt)
	})
	if query.Limit > 0 && len(out) > query.Limit {
		out = out[:query.Limit]
	}
	return out
}

func SummarizeEntries(entries []LedgerEntry) Summary {
	var out Summary
	for _, entry := range entries {
		out.Total++
		out.TotalDurationMs += entry.DurationMs
		switch strings.ToLower(strings.TrimSpace(entry.Status)) {
		case "applied":
			out.Applied++
		case "skipped":
			out.Skipped++
		case "failed":
			out.Failed++
		}
	}
	return out
}

func normalizeStatuses(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.ToLower(strings.TrimSpace(value))
		if trimmed == "" {
			continue
		}
		out[trimmed] = struct{}{}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func statusAllowed(status string, allowed map[string]struct{}) bool {
	_, ok := allowed[strings.ToLower(strings.TrimSpace(status))]
	return ok
}
