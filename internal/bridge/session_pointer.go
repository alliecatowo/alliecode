package bridge

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type sessionPointerPayload struct {
	SessionID string `json:"session_id"`
}

// SessionPointer references a session in a stable encoded form.
type SessionPointer struct {
	SessionID string
}

// EncodeSessionPointer converts a session ID into a bridge pointer.
func EncodeSessionPointer(sessionID string) (string, error) {
	normalized, err := NormalizeSessionID(sessionID)
	if err != nil {
		return "", err
	}
	payload := sessionPointerPayload{SessionID: normalized}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal pointer payload: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return pointerPrefix + token, nil
}

// DecodeSessionPointer parses a bridge pointer.
func DecodeSessionPointer(pointer string) (SessionPointer, error) {
	trimmed := strings.TrimSpace(pointer)
	if !strings.HasPrefix(trimmed, pointerPrefix) {
		return SessionPointer{}, fmt.Errorf("invalid session pointer %q", pointer)
	}
	token := strings.TrimPrefix(trimmed, pointerPrefix)
	if token == "" {
		return SessionPointer{}, fmt.Errorf("invalid session pointer %q", pointer)
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return SessionPointer{}, fmt.Errorf("decode session pointer: %w", err)
	}
	var payload sessionPointerPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return SessionPointer{}, fmt.Errorf("unmarshal session pointer: %w", err)
	}
	normalized, err := NormalizeSessionID(payload.SessionID)
	if err != nil {
		return SessionPointer{}, err
	}
	return SessionPointer{SessionID: normalized}, nil
}
