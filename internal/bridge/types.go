package bridge

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	legacySessionPrefix = "session:"
	pointerPrefix       = "bridge://session/"
	compatSessionPrefix = "session_"
	infraSessionPrefix  = "cse_"
)

var sessionIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)

var (
	errSessionIDEmpty            = errors.New("session ID cannot be empty")
	errEnvelopeFromEmpty         = errors.New("envelope from cannot be empty")
	errEnvelopeToEmpty           = errors.New("envelope to cannot be empty")
	errEnvelopeKindEmpty         = errors.New("envelope kind cannot be empty")
	errEnvelopeIDWhitespace      = errors.New("envelope id cannot include leading or trailing whitespace")
	errEnvelopeCorrelationSpaces = errors.New("envelope correlation_id cannot include leading or trailing whitespace")
	errEnvelopeCreatedAtZero     = errors.New("envelope created_at cannot be zero")
)

var validEnvelopeKinds = map[MessageKind]struct{}{
	MessageKindData:                     {},
	MessageKindAck:                      {},
	MessageKindEvent:                    {},
	MessageKindRequest:                  {},
	MessageKindResponse:                 {},
	MessageKindCancel:                   {},
	MessageKind(ProtocolTypePermission): {},
	MessageKind(ProtocolTypeHeartbeat):  {},
}

// MessageKind labels the purpose of one bridge envelope.
type MessageKind string

const (
	MessageKindData     MessageKind = "data"
	MessageKindAck      MessageKind = "ack"
	MessageKindEvent    MessageKind = "event"
	MessageKindRequest  MessageKind = "request"
	MessageKindResponse MessageKind = "response"
	MessageKindCancel   MessageKind = "cancel"
)

// Envelope is the transport-agnostic bridge message unit.
type Envelope struct {
	ID            string      `json:"id"`
	SessionID     string      `json:"session_id"`
	From          string      `json:"from"`
	To            string      `json:"to"`
	Kind          MessageKind `json:"kind"`
	Payload       []byte      `json:"payload,omitempty"`
	CorrelationID string      `json:"correlation_id,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
}

// SessionRecord holds local metadata for one bridge session.
type SessionRecord struct {
	SessionID      string    `json:"session_id"`
	Pointer        string    `json:"pointer"`
	WorkSecretHash string    `json:"work_secret_hash"`
	CreatedAt      time.Time `json:"created_at"`
}

// NormalizeSessionID accepts compatibility formats and validates IDs.
func NormalizeSessionID(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errSessionIDEmpty
	}

	if strings.HasPrefix(trimmed, pointerPrefix) {
		ptr, err := DecodeSessionPointer(trimmed)
		if err != nil {
			return "", err
		}
		return ptr.SessionID, nil
	}

	if strings.HasPrefix(trimmed, legacySessionPrefix) {
		trimmed = strings.TrimPrefix(trimmed, legacySessionPrefix)
		if trimmed == "" {
			return "", errSessionIDEmpty
		}
	}

	if !sessionIDPattern.MatchString(trimmed) {
		return "", fmt.Errorf("invalid session ID %q", raw)
	}

	return trimmed, nil
}

// ToCompatSessionID rewrites infra-tagged IDs for compat surfaces.
func ToCompatSessionID(raw string) string {
	if !strings.HasPrefix(raw, infraSessionPrefix) {
		return raw
	}
	return compatSessionPrefix + strings.TrimPrefix(raw, infraSessionPrefix)
}

// ToInfraSessionID rewrites compat-tagged IDs for infra surfaces.
func ToInfraSessionID(raw string) string {
	if !strings.HasPrefix(raw, compatSessionPrefix) {
		return raw
	}
	return infraSessionPrefix + strings.TrimPrefix(raw, compatSessionPrefix)
}

// ValidateEnvelope checks required routing fields.
func ValidateEnvelope(env Envelope) error {
	if strings.TrimSpace(env.SessionID) != env.SessionID {
		return fmt.Errorf("envelope session_id %q has leading or trailing whitespace", env.SessionID)
	}
	if _, err := NormalizeSessionID(env.SessionID); err != nil {
		return fmt.Errorf("envelope session_id invalid: %w", err)
	}
	if strings.TrimSpace(env.From) == "" {
		return errEnvelopeFromEmpty
	}
	if strings.TrimSpace(env.From) != env.From {
		return errors.New("envelope from cannot include leading or trailing whitespace")
	}
	if strings.TrimSpace(env.To) == "" {
		return errEnvelopeToEmpty
	}
	if strings.TrimSpace(env.To) != env.To {
		return errors.New("envelope to cannot include leading or trailing whitespace")
	}
	if env.Kind == "" {
		return errEnvelopeKindEmpty
	}
	if strings.TrimSpace(string(env.Kind)) != string(env.Kind) {
		return errors.New("envelope kind cannot include leading or trailing whitespace")
	}
	if _, ok := validEnvelopeKinds[env.Kind]; !ok {
		return fmt.Errorf("unsupported envelope kind %q", env.Kind)
	}
	if strings.TrimSpace(env.ID) != env.ID {
		return errEnvelopeIDWhitespace
	}
	if strings.TrimSpace(env.CorrelationID) != env.CorrelationID {
		return errEnvelopeCorrelationSpaces
	}
	if env.CreatedAt.IsZero() {
		return errEnvelopeCreatedAtZero
	}
	return nil
}
