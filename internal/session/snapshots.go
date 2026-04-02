package session

import (
	"sort"
	"strings"
)

type Snapshot struct {
	SessionID       string `json:"session_id"`
	ProjectPath     string `json:"project_path,omitempty"`
	Provider        string `json:"provider,omitempty"`
	Model           string `json:"model,omitempty"`
	RuntimeSurface  string `json:"runtime_surface,omitempty"`
	CommandSurface  string `json:"command_surface,omitempty"`
	MessageCount    int    `json:"message_count,omitempty"`
	EventCount      int    `json:"event_count,omitempty"`
	LastMessageRole string `json:"last_message_role,omitempty"`
	LastEventUnix   int64  `json:"last_event_unix,omitempty"`
}

type SnapshotQuery struct {
	ProjectContains string
	Provider        string
	Model           string
	RuntimeSurface  string
	CommandSurface  string
	LastRole        string
	MinMessages     int
	Limit           int
}

func BuildSnapshots(items []Transcript, query SnapshotQuery) []Snapshot {
	if len(items) == 0 {
		return nil
	}
	projectContains := strings.ToLower(strings.TrimSpace(query.ProjectContains))
	provider := strings.ToLower(strings.TrimSpace(query.Provider))
	model := strings.ToLower(strings.TrimSpace(query.Model))
	runtimeSurface := strings.ToLower(strings.TrimSpace(query.RuntimeSurface))
	commandSurface := strings.ToLower(strings.TrimSpace(query.CommandSurface))
	lastRole := strings.ToLower(strings.TrimSpace(query.LastRole))

	out := make([]Snapshot, 0, len(items))
	for _, item := range items {
		if projectContains != "" && !strings.Contains(strings.ToLower(strings.TrimSpace(item.ProjectPath)), projectContains) {
			continue
		}
		if provider != "" && strings.ToLower(strings.TrimSpace(item.Provider)) != provider {
			continue
		}
		if model != "" && strings.ToLower(strings.TrimSpace(item.Model)) != model {
			continue
		}
		if runtimeSurface != "" && strings.ToLower(strings.TrimSpace(item.RuntimeSurface)) != runtimeSurface {
			continue
		}
		if commandSurface != "" && strings.ToLower(strings.TrimSpace(item.CommandSurface)) != commandSurface {
			continue
		}
		if lastRole != "" && strings.ToLower(strings.TrimSpace(item.LastMessageRole)) != lastRole {
			continue
		}
		if query.MinMessages > 0 && item.MessageCount < query.MinMessages {
			continue
		}
		out = append(out, Snapshot{
			SessionID:       item.SessionID,
			ProjectPath:     item.ProjectPath,
			Provider:        item.Provider,
			Model:           item.Model,
			RuntimeSurface:  item.RuntimeSurface,
			CommandSurface:  item.CommandSurface,
			MessageCount:    item.MessageCount,
			EventCount:      item.EventCount,
			LastMessageRole: item.LastMessageRole,
			LastEventUnix:   item.LastEventAt.Unix(),
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].LastEventUnix != out[j].LastEventUnix {
			return out[i].LastEventUnix > out[j].LastEventUnix
		}
		return out[i].SessionID < out[j].SessionID
	})
	if query.Limit > 0 && len(out) > query.Limit {
		out = out[:query.Limit]
	}
	return out
}
