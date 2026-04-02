package state

import "strings"

type SessionMetadataRegistry struct {
	Sessions          map[string]SessionMetadata `json:"sessions,omitempty"`
	LastSessionID     string                     `json:"last_session_id,omitempty"`
	LastUpdatedAt     int64                      `json:"last_updated_at,omitempty"`
	TotalSessions     int                        `json:"total_sessions,omitempty"`
	TotalMessageCount int                        `json:"total_message_count,omitempty"`
}

type SessionMetadata struct {
	SessionID      string `json:"session_id,omitempty"`
	SessionPath    string `json:"session_path,omitempty"`
	ProjectPath    string `json:"project_path,omitempty"`
	Title          string `json:"title,omitempty"`
	Summary        string `json:"summary,omitempty"`
	UpdatedAt      int64  `json:"updated_at,omitempty"`
	CreatedAt      int64  `json:"created_at,omitempty"`
	MessageCount   int    `json:"message_count,omitempty"`
	EventCount     int    `json:"event_count,omitempty"`
	LastRole       string `json:"last_role,omitempty"`
	RuntimeSurface string `json:"runtime_surface,omitempty"`
	CommandSurface string `json:"command_surface,omitempty"`
}

type SessionMetadataUpdate struct {
	SessionID      string
	SessionPath    string
	ProjectPath    string
	Title          string
	Summary        string
	UpdatedAt      int64
	CreatedAt      int64
	MessageCount   int
	EventCount     int
	LastRole       string
	RuntimeSurface string
	CommandSurface string
}

func ReadSessionMetadataRegistry(path string) (SessionMetadataRegistry, error) {
	state := SessionMetadataRegistry{Sessions: make(map[string]SessionMetadata)}
	if err := readJSONState(path, "session metadata", &state); err != nil {
		return SessionMetadataRegistry{}, err
	}
	return normalizeSessionMetadataRegistry(state), nil
}

func WriteSessionMetadataRegistry(path string, state SessionMetadataRegistry) error {
	state = normalizeSessionMetadataRegistry(state)
	return writeJSONState(path, "session metadata", state)
}

func normalizeSessionMetadataRegistry(state SessionMetadataRegistry) SessionMetadataRegistry {
	state.LastSessionID = strings.TrimSpace(state.LastSessionID)
	if state.LastUpdatedAt < 0 {
		state.LastUpdatedAt = 0
	}
	if state.TotalSessions < 0 {
		state.TotalSessions = 0
	}
	if state.TotalMessageCount < 0 {
		state.TotalMessageCount = 0
	}
	cleaned := make(map[string]SessionMetadata)
	var totalMessageCount int
	var maxUpdatedAt int64
	var lastSessionID string
	for sessionID, metadata := range state.Sessions {
		sessionID = strings.TrimSpace(sessionID)
		if sessionID == "" {
			continue
		}
		metadata.SessionID = strings.TrimSpace(metadata.SessionID)
		if metadata.SessionID == "" {
			metadata.SessionID = sessionID
		}
		metadata.SessionPath = strings.TrimSpace(metadata.SessionPath)
		metadata.ProjectPath = strings.TrimSpace(metadata.ProjectPath)
		metadata.Title = strings.TrimSpace(metadata.Title)
		metadata.Summary = strings.TrimSpace(metadata.Summary)
		metadata.LastRole = strings.TrimSpace(metadata.LastRole)
		metadata.RuntimeSurface = strings.TrimSpace(metadata.RuntimeSurface)
		metadata.CommandSurface = strings.TrimSpace(metadata.CommandSurface)
		if metadata.UpdatedAt < 0 {
			metadata.UpdatedAt = 0
		}
		if metadata.CreatedAt < 0 {
			metadata.CreatedAt = 0
		}
		if metadata.CreatedAt > 0 && metadata.UpdatedAt > 0 && metadata.CreatedAt > metadata.UpdatedAt {
			metadata.CreatedAt = metadata.UpdatedAt
		}
		if metadata.MessageCount < 0 {
			metadata.MessageCount = 0
		}
		if metadata.EventCount < 0 {
			metadata.EventCount = 0
		}
		totalMessageCount += metadata.MessageCount
		if metadata.UpdatedAt >= maxUpdatedAt {
			maxUpdatedAt = metadata.UpdatedAt
			lastSessionID = metadata.SessionID
		}
		cleaned[sessionID] = metadata
	}
	state.Sessions = cleaned
	state.TotalSessions = len(cleaned)
	if state.TotalMessageCount == 0 {
		state.TotalMessageCount = totalMessageCount
	}
	if state.LastUpdatedAt == 0 {
		state.LastUpdatedAt = maxUpdatedAt
	}
	if state.LastSessionID == "" {
		state.LastSessionID = lastSessionID
	}
	return state
}

func UpsertSessionMetadata(path string, update SessionMetadataUpdate) (SessionMetadataRegistry, error) {
	registry, err := ReadSessionMetadataRegistry(path)
	if err != nil {
		return SessionMetadataRegistry{}, err
	}

	sessionID := strings.TrimSpace(update.SessionID)
	if sessionID == "" {
		return registry, nil
	}

	entry := registry.Sessions[sessionID]
	if entry.SessionID == "" {
		entry.SessionID = sessionID
	}
	if v := strings.TrimSpace(update.SessionPath); v != "" {
		entry.SessionPath = v
	}
	if v := strings.TrimSpace(update.ProjectPath); v != "" {
		entry.ProjectPath = v
	}
	if v := strings.TrimSpace(update.Title); v != "" {
		entry.Title = v
	}
	if v := strings.TrimSpace(update.Summary); v != "" {
		entry.Summary = v
	}
	if v := strings.TrimSpace(update.LastRole); v != "" {
		entry.LastRole = v
	}
	if v := strings.TrimSpace(update.RuntimeSurface); v != "" {
		entry.RuntimeSurface = v
	}
	if v := strings.TrimSpace(update.CommandSurface); v != "" {
		entry.CommandSurface = v
	}
	if update.CreatedAt > 0 && (entry.CreatedAt == 0 || update.CreatedAt < entry.CreatedAt) {
		entry.CreatedAt = update.CreatedAt
	}
	if update.UpdatedAt > 0 {
		entry.UpdatedAt = update.UpdatedAt
	}
	if update.MessageCount >= 0 {
		entry.MessageCount = update.MessageCount
	}
	if update.EventCount >= 0 {
		entry.EventCount = update.EventCount
	}

	registry.Sessions[sessionID] = entry
	registry.LastSessionID = sessionID
	if update.UpdatedAt > 0 {
		registry.LastUpdatedAt = update.UpdatedAt
	}
	registry.TotalSessions = len(registry.Sessions)
	registry.TotalMessageCount = aggregateSessionMessageCount(registry.Sessions)

	if err := WriteSessionMetadataRegistry(path, registry); err != nil {
		return SessionMetadataRegistry{}, err
	}
	return registry, nil
}

func aggregateSessionMessageCount(items map[string]SessionMetadata) int {
	total := 0
	for _, item := range items {
		total += item.MessageCount
	}
	if total < 0 {
		return 0
	}
	return total
}
