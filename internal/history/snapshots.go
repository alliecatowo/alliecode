package history

import (
	"sort"
	"strings"
)

type Snapshot struct {
	SessionID      string `json:"session_id,omitempty"`
	Project        string `json:"project,omitempty"`
	Display        string `json:"display,omitempty"`
	Type           string `json:"type,omitempty"`
	TimestampUnix  int64  `json:"timestamp_unix,omitempty"`
	RuntimeSurface string `json:"runtime_surface,omitempty"`
	CommandSurface string `json:"command_surface,omitempty"`
	Provider       string `json:"provider,omitempty"`
	Model          string `json:"model,omitempty"`
	TagCount       int    `json:"tag_count,omitempty"`
}

type SnapshotQuery struct {
	ProjectContains string
	SessionContains string
	DisplayContains string
	Provider        string
	Model           string
	Limit           int
}

func BuildSnapshots(events []Event, query SnapshotQuery) []Snapshot {
	if len(events) == 0 {
		return nil
	}
	projectContains := strings.ToLower(strings.TrimSpace(query.ProjectContains))
	sessionContains := strings.ToLower(strings.TrimSpace(query.SessionContains))
	displayContains := strings.ToLower(strings.TrimSpace(query.DisplayContains))
	provider := strings.ToLower(strings.TrimSpace(query.Provider))
	model := strings.ToLower(strings.TrimSpace(query.Model))

	out := make([]Snapshot, 0, len(events))
	for _, ev := range events {
		display := strings.TrimSpace(eventDisplay(ev))
		if projectContains != "" && !strings.Contains(strings.ToLower(strings.TrimSpace(ev.Project)), projectContains) {
			continue
		}
		if sessionContains != "" && !strings.Contains(strings.ToLower(strings.TrimSpace(ev.SessionID)), sessionContains) {
			continue
		}
		if displayContains != "" && !strings.Contains(strings.ToLower(display), displayContains) {
			continue
		}
		if provider != "" && strings.ToLower(strings.TrimSpace(ev.Provider)) != provider {
			continue
		}
		if model != "" && strings.ToLower(strings.TrimSpace(ev.Model)) != model {
			continue
		}
		out = append(out, Snapshot{
			SessionID:      strings.TrimSpace(ev.SessionID),
			Project:        strings.TrimSpace(ev.Project),
			Display:        display,
			Type:           string(ev.Type),
			TimestampUnix:  ev.Timestamp.Unix(),
			RuntimeSurface: strings.TrimSpace(ev.RuntimeSurface),
			CommandSurface: strings.TrimSpace(ev.CommandSurface),
			Provider:       strings.TrimSpace(ev.Provider),
			Model:          strings.TrimSpace(ev.Model),
			TagCount:       len(ev.Tags),
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].TimestampUnix != out[j].TimestampUnix {
			return out[i].TimestampUnix > out[j].TimestampUnix
		}
		return out[i].SessionID < out[j].SessionID
	})

	if query.Limit > 0 && len(out) > query.Limit {
		out = out[:query.Limit]
	}
	return out
}
