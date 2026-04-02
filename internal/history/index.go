package history

import (
	"sort"
	"strings"
	"time"
)

type IndexRecord struct {
	SessionID      string    `json:"session_id"`
	Project        string    `json:"project,omitempty"`
	Display        string    `json:"display,omitempty"`
	Type           EventType `json:"type"`
	Timestamp      time.Time `json:"timestamp"`
	RuntimeSurface string    `json:"runtime_surface,omitempty"`
	CommandSurface string    `json:"command_surface,omitempty"`
	Provider       string    `json:"provider,omitempty"`
	Model          string    `json:"model,omitempty"`
}

type IndexQuery struct {
	ProjectContains string
	SessionContains string
	DisplayContains string
	RuntimeSurface  string
	CommandSurface  string
	TagContains     string
	Type            EventType
	Provider        string
	Model           string
	Limit           int
}

func BuildIndex(events []Event, query IndexQuery) []IndexRecord {
	if len(events) == 0 {
		return nil
	}
	projectFilter := strings.ToLower(strings.TrimSpace(query.ProjectContains))
	sessionFilter := strings.ToLower(strings.TrimSpace(query.SessionContains))
	displayFilter := strings.ToLower(strings.TrimSpace(query.DisplayContains))
	providerFilter := strings.ToLower(strings.TrimSpace(query.Provider))
	modelFilter := strings.ToLower(strings.TrimSpace(query.Model))
	runtimeSurfaceFilter := strings.ToLower(strings.TrimSpace(query.RuntimeSurface))
	commandSurfaceFilter := strings.ToLower(strings.TrimSpace(query.CommandSurface))
	tagContainsFilter := strings.ToLower(strings.TrimSpace(query.TagContains))

	out := make([]IndexRecord, 0, len(events))
	for _, ev := range events {
		display := strings.TrimSpace(eventDisplay(ev))
		if query.Type != "" && ev.Type != query.Type {
			continue
		}
		if projectFilter != "" && !strings.Contains(strings.ToLower(strings.TrimSpace(ev.Project)), projectFilter) {
			continue
		}
		if sessionFilter != "" && !strings.Contains(strings.ToLower(strings.TrimSpace(ev.SessionID)), sessionFilter) {
			continue
		}
		if displayFilter != "" && !strings.Contains(strings.ToLower(display), displayFilter) {
			continue
		}
		if providerFilter != "" && strings.ToLower(strings.TrimSpace(ev.Provider)) != providerFilter {
			continue
		}
		if modelFilter != "" && strings.ToLower(strings.TrimSpace(ev.Model)) != modelFilter {
			continue
		}
		if runtimeSurfaceFilter != "" && strings.ToLower(strings.TrimSpace(ev.RuntimeSurface)) != runtimeSurfaceFilter {
			continue
		}
		if commandSurfaceFilter != "" && strings.ToLower(strings.TrimSpace(ev.CommandSurface)) != commandSurfaceFilter {
			continue
		}
		if tagContainsFilter != "" && !eventHasTag(ev.Tags, tagContainsFilter) {
			continue
		}
		out = append(out, IndexRecord{
			SessionID:      strings.TrimSpace(ev.SessionID),
			Project:        strings.TrimSpace(ev.Project),
			Display:        display,
			Type:           ev.Type,
			Timestamp:      ev.Timestamp,
			RuntimeSurface: strings.TrimSpace(ev.RuntimeSurface),
			CommandSurface: strings.TrimSpace(ev.CommandSurface),
			Provider:       strings.TrimSpace(ev.Provider),
			Model:          strings.TrimSpace(ev.Model),
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Timestamp.After(out[j].Timestamp)
	})
	if query.Limit > 0 && len(out) > query.Limit {
		out = out[:query.Limit]
	}
	return out
}

func eventHasTag(tags []string, tagContainsFilter string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(strings.TrimSpace(tag)), tagContainsFilter) {
			return true
		}
	}
	return false
}
