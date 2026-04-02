package bridge

import (
	"errors"
	"sync"
	"time"
)

// StateManager tracks local bridge session metadata.
type StateManager struct {
	mu       sync.RWMutex
	sessions map[string]SessionRecord
}

// NewStateManager constructs an empty bridge state manager.
func NewStateManager() *StateManager {
	return &StateManager{sessions: make(map[string]SessionRecord)}
}

// AddSession registers a new session and returns the stored record.
func (m *StateManager) AddSession(sessionID, workSecret string) (SessionRecord, error) {
	normalized, err := NormalizeSessionID(sessionID)
	if err != nil {
		return SessionRecord{}, err
	}
	pointer, err := EncodeSessionPointer(normalized)
	if err != nil {
		return SessionRecord{}, err
	}
	hash, err := WorkSecretHash(workSecret)
	if err != nil {
		return SessionRecord{}, err
	}

	record := SessionRecord{
		SessionID:      normalized,
		Pointer:        pointer,
		WorkSecretHash: hash,
		CreatedAt:      time.Now().UTC(),
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.sessions[normalized]; exists {
		return SessionRecord{}, errors.New("session already exists")
	}
	m.sessions[normalized] = record
	return record, nil
}

// GetSession returns a record by session ID.
func (m *StateManager) GetSession(sessionID string) (SessionRecord, bool) {
	normalized, err := NormalizeSessionID(sessionID)
	if err != nil {
		return SessionRecord{}, false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	record, ok := m.sessions[normalized]
	return record, ok
}

// DeleteSession removes a session from local state.
func (m *StateManager) DeleteSession(sessionID string) {
	normalized, err := NormalizeSessionID(sessionID)
	if err != nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, normalized)
}

// ListSessions returns all session records in arbitrary order.
func (m *StateManager) ListSessions() []SessionRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]SessionRecord, 0, len(m.sessions))
	for _, record := range m.sessions {
		out = append(out, record)
	}
	return out
}
