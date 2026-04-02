package remote

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// SessionState tracks the lifecycle of one remote session.
type SessionState string

const (
	SessionStateConnecting   SessionState = "connecting"
	SessionStateActive       SessionState = "active"
	SessionStateReconnecting SessionState = "reconnecting"
	SessionStateClosed       SessionState = "closed"
)

// TransportState tracks lower-level transport connectivity lifecycle.
type TransportState string

const (
	TransportStateConnecting   TransportState = "connecting"
	TransportStateConnected    TransportState = "connected"
	TransportStateReconnecting TransportState = "reconnecting"
	TransportStateClosed       TransportState = "closed"
)

// TransportErrorClass groups transport failures by reconnect behavior.
type TransportErrorClass string

const (
	TransportErrorClassTransient TransportErrorClass = "transient"
	TransportErrorClassPermanent TransportErrorClass = "permanent"
	TransportErrorClassAuth      TransportErrorClass = "auth"
	TransportErrorClassTimeout   TransportErrorClass = "timeout"
	TransportErrorClassCanceled  TransportErrorClass = "canceled"
	TransportErrorClassUnknown   TransportErrorClass = "unknown"
)

// ReconnectReason describes why reconnect flow started or ended.
type ReconnectReason string

const (
	ReconnectReasonNone                ReconnectReason = "none"
	ReconnectReasonDisconnect          ReconnectReason = "disconnect"
	ReconnectReasonConnectFailure      ReconnectReason = "connect_failure"
	ReconnectReasonHeartbeatExpired    ReconnectReason = "heartbeat_expired"
	ReconnectReasonReconnectFailed     ReconnectReason = "reconnect_failed"
	ReconnectReasonReconnectMaxRetries ReconnectReason = "reconnect_max_retries"
	ReconnectReasonClosed              ReconnectReason = "closed"
)

// ReconnectCause captures a classified reconnect reason with details.
type ReconnectCause struct {
	Reason         ReconnectReason     `json:"reason"`
	ErrorClass     TransportErrorClass `json:"error_class"`
	Detail         string              `json:"detail,omitempty"`
	ReconnectCount int                 `json:"reconnect_count,omitempty"`
}

// HealthSummary reports current manager and reconnect health details.
type HealthSummary struct {
	State          SessionState   `json:"state"`
	TransportState TransportState `json:"transport_state"`
	LastCause      ReconnectCause `json:"last_cause"`
	ReconnectCount int            `json:"reconnect_count"`
	Reconnecting   bool           `json:"reconnecting"`
	RequestStats   RequestStats   `json:"request_stats"`
}

// MessageKind identifies protocol-level remote message classes.
type MessageKind string

const (
	MessageKindProviderEvent MessageKind = "provider_event"
	MessageKindToolEvent     MessageKind = "tool_event"
	MessageKindHeartbeat     MessageKind = "heartbeat"
	MessageKindRequest       MessageKind = "request"
	MessageKindResponse      MessageKind = "response"
	MessageKindCancel        MessageKind = "cancel"
)

// RemoteMessage is the transport payload exchanged for remote sessions.
type RemoteMessage struct {
	SessionID string          `json:"session_id"`
	Kind      MessageKind     `json:"kind"`
	Name      string          `json:"name,omitempty"`
	Seq       int64           `json:"seq"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	SentAt    time.Time       `json:"sent_at"`
}

// Validate ensures protocol messages include required routing fields.
func (m RemoteMessage) Validate() error {
	if strings.TrimSpace(m.SessionID) == "" {
		return errors.New("session_id cannot be empty")
	}
	if m.Kind == "" {
		return errors.New("kind cannot be empty")
	}
	if m.Seq <= 0 {
		return fmt.Errorf("seq must be positive: %d", m.Seq)
	}
	if m.SentAt.IsZero() {
		return errors.New("sent_at cannot be zero")
	}
	return nil
}

func isHeartbeat(msg RemoteMessage) bool {
	return msg.Kind == MessageKindHeartbeat
}
