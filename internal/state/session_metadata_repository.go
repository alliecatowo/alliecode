package state

import (
	"fmt"
	"sort"
)

type SessionMetadataRepository struct {
	path string
}

type SessionMetadataQuery struct {
	ProjectContains   string
	SessionContains   string
	TitleContains     string
	RuntimeSurface    string
	CommandSurface    string
	UpdatedAfterUnix  int64
	UpdatedBeforeUnix int64
	MinMessages       int
	Limit             int
	SortByMessages    bool
}

type SessionMetadataIndexQuery struct {
	ProjectContains string
	SessionContains string
	Role            string
	RuntimeSurface  string
	CommandSurface  string
	MinMessages     int
	MinEvents       int
	UpdatedAfter    int64
	Limit           int
}

type SessionMetadataIndexRecord struct {
	SessionID         string `json:"session_id"`
	ProjectPath       string `json:"project_path,omitempty"`
	UpdatedAt         int64  `json:"updated_at,omitempty"`
	MessageCount      int    `json:"message_count,omitempty"`
	EventCount        int    `json:"event_count,omitempty"`
	LastRole          string `json:"last_role,omitempty"`
	RuntimeSurface    string `json:"runtime_surface,omitempty"`
	CommandSurface    string `json:"command_surface,omitempty"`
	MessageEventRatio string `json:"message_event_ratio,omitempty"`
}

type SessionMetadataSnapshot struct {
	SessionID string          `json:"session_id"`
	Metadata  SessionMetadata `json:"metadata"`
}

func NewSessionMetadataRepository(path string) *SessionMetadataRepository {
	return &SessionMetadataRepository{path: path}
}

func (r *SessionMetadataRepository) Path() string {
	if r == nil {
		return ""
	}
	return r.path
}

func (r *SessionMetadataRepository) Get() (SessionMetadataRegistry, error) {
	return ReadSessionMetadataRegistry(r.path)
}

func (r *SessionMetadataRepository) Save(registry SessionMetadataRegistry) error {
	return WriteSessionMetadataRegistry(r.path, registry)
}

func (r *SessionMetadataRepository) Upsert(update SessionMetadataUpdate) (SessionMetadataRegistry, error) {
	return UpsertSessionMetadata(r.path, update)
}

func (r *SessionMetadataRepository) Query(query SessionMetadataQuery) ([]SessionMetadataSnapshot, error) {
	registry, err := r.Get()
	if err != nil {
		return nil, err
	}

	projectContains := normalizeContains(query.ProjectContains)
	sessionContains := normalizeContains(query.SessionContains)
	titleContains := normalizeContains(query.TitleContains)
	runtimeSurface := normalizeContains(query.RuntimeSurface)
	commandSurface := normalizeContains(query.CommandSurface)

	items := make([]SessionMetadataSnapshot, 0, len(registry.Sessions))
	for id, metadata := range registry.Sessions {
		if !matchesContains(metadata.ProjectPath, projectContains) {
			continue
		}
		if !matchesContains(id, sessionContains) {
			continue
		}
		if !matchesContains(metadata.Title, titleContains) {
			continue
		}
		if runtimeSurface != "" && normalizeContains(metadata.RuntimeSurface) != runtimeSurface {
			continue
		}
		if commandSurface != "" && normalizeContains(metadata.CommandSurface) != commandSurface {
			continue
		}
		if query.UpdatedAfterUnix > 0 && metadata.UpdatedAt < query.UpdatedAfterUnix {
			continue
		}
		if query.UpdatedBeforeUnix > 0 && metadata.UpdatedAt > query.UpdatedBeforeUnix {
			continue
		}
		if query.MinMessages > 0 && metadata.MessageCount < query.MinMessages {
			continue
		}
		items = append(items, SessionMetadataSnapshot{SessionID: id, Metadata: metadata})
	}

	if query.SortByMessages {
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].Metadata.MessageCount != items[j].Metadata.MessageCount {
				return items[i].Metadata.MessageCount > items[j].Metadata.MessageCount
			}
			if items[i].Metadata.UpdatedAt != items[j].Metadata.UpdatedAt {
				return items[i].Metadata.UpdatedAt > items[j].Metadata.UpdatedAt
			}
			return items[i].SessionID < items[j].SessionID
		})
	} else {
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].Metadata.UpdatedAt != items[j].Metadata.UpdatedAt {
				return items[i].Metadata.UpdatedAt > items[j].Metadata.UpdatedAt
			}
			return items[i].SessionID < items[j].SessionID
		})
	}

	limit := clampLimit(query.Limit, len(items))
	return items[:limit], nil
}

func (r *SessionMetadataRepository) QueryIndex(query SessionMetadataIndexQuery) ([]SessionMetadataIndexRecord, error) {
	registry, err := r.Get()
	if err != nil {
		return nil, err
	}
	return BuildSessionMetadataIndex(registry, query), nil
}

func BuildSessionMetadataIndex(registry SessionMetadataRegistry, query SessionMetadataIndexQuery) []SessionMetadataIndexRecord {
	if len(registry.Sessions) == 0 {
		return nil
	}
	projectContains := normalizeContains(query.ProjectContains)
	sessionContains := normalizeContains(query.SessionContains)
	role := normalizeEquals(query.Role)
	runtimeSurface := normalizeEquals(query.RuntimeSurface)
	commandSurface := normalizeEquals(query.CommandSurface)

	out := make([]SessionMetadataIndexRecord, 0, len(registry.Sessions))
	for sessionID, metadata := range registry.Sessions {
		if !matchesContains(metadata.ProjectPath, projectContains) {
			continue
		}
		if !matchesContains(sessionID, sessionContains) {
			continue
		}
		if !matchesEquals(metadata.LastRole, role) {
			continue
		}
		if !matchesEquals(metadata.RuntimeSurface, runtimeSurface) {
			continue
		}
		if !matchesEquals(metadata.CommandSurface, commandSurface) {
			continue
		}
		if query.MinMessages > 0 && metadata.MessageCount < query.MinMessages {
			continue
		}
		if query.MinEvents > 0 && metadata.EventCount < query.MinEvents {
			continue
		}
		if query.UpdatedAfter > 0 && metadata.UpdatedAt < query.UpdatedAfter {
			continue
		}
		ratio := "0/0"
		if metadata.EventCount > 0 {
			ratio = normalizeRatio(metadata.MessageCount, metadata.EventCount)
		}
		out = append(out, SessionMetadataIndexRecord{
			SessionID:         sessionID,
			ProjectPath:       metadata.ProjectPath,
			UpdatedAt:         metadata.UpdatedAt,
			MessageCount:      metadata.MessageCount,
			EventCount:        metadata.EventCount,
			LastRole:          metadata.LastRole,
			RuntimeSurface:    metadata.RuntimeSurface,
			CommandSurface:    metadata.CommandSurface,
			MessageEventRatio: ratio,
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].UpdatedAt != out[j].UpdatedAt {
			return out[i].UpdatedAt > out[j].UpdatedAt
		}
		if out[i].MessageCount != out[j].MessageCount {
			return out[i].MessageCount > out[j].MessageCount
		}
		return out[i].SessionID < out[j].SessionID
	})
	limit := clampLimit(query.Limit, len(out))
	return out[:limit]
}

func normalizeRatio(messages, events int) string {
	if events <= 0 {
		return "0/0"
	}
	if messages < 0 {
		messages = 0
	}
	return fmt.Sprintf("%d/%d", messages, events)
}
