package remote

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

var errReconnectAbortedClosed = errors.New("reconnect aborted: session closed")

// SessionManager drives state transitions for one remote session.
type SessionManager struct {
	transport Transport
	backoff   *Backoff
	heartbeat *HeartbeatMonitor
	policy    RetryPolicy
	requests  *RequestCorrelator

	mu             sync.Mutex
	state          SessionState
	transportState TransportState
	session        string
	attempt        int
	reconnecting   bool
	closeCh        chan struct{}
	lastCause      ReconnectCause
}

// NewSessionManager constructs a session manager with defaults.
func NewSessionManager(transport Transport, sessionID string, policy RetryPolicy, now time.Time) (*SessionManager, error) {
	if transport == nil {
		return nil, errors.New("transport cannot be nil")
	}
	if sessionID == "" {
		return nil, errors.New("sessionID cannot be empty")
	}
	if policy == (RetryPolicy{}) {
		policy = DefaultRetryPolicy()
	}

	return &SessionManager{
		transport:      transport,
		backoff:        NewBackoff(policy, 1),
		heartbeat:      NewHeartbeatMonitor(maxDuration(3*policy.MaxDelay, 3*time.Second), now),
		policy:         policy,
		requests:       NewRequestCorrelator(),
		state:          SessionStateClosed,
		transportState: TransportStateClosed,
		session:        sessionID,
		closeCh:        make(chan struct{}),
		lastCause:      ReconnectCause{Reason: ReconnectReasonNone, ErrorClass: TransportErrorClassUnknown},
	}, nil
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

// State returns the current manager state.
func (m *SessionManager) State() SessionState {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

// TransportState returns the current manager transport state.
func (m *SessionManager) TransportState() TransportState {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.transportState
}

// LastReconnectCause returns the last classified reconnect cause.
func (m *SessionManager) LastReconnectCause() ReconnectCause {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastCause
}

// HealthSummary returns current state and reconnect telemetry.
func (m *SessionManager) HealthSummary() HealthSummary {
	m.mu.Lock()
	defer m.mu.Unlock()
	stats := RequestStats{}
	if m.requests != nil {
		stats = m.requests.Stats()
	}
	return HealthSummary{
		State:          m.state,
		TransportState: m.transportState,
		LastCause:      m.lastCause,
		ReconnectCount: m.attempt,
		Reconnecting:   m.reconnecting,
		RequestStats:   stats,
	}
}

// ObserveMessage records request/response/cancel lifecycle transitions.
func (m *SessionManager) ObserveMessage(msg RemoteMessage) error {
	if m == nil || m.requests == nil {
		return errors.New("session manager request correlator is not initialized")
	}
	return m.requests.Observe(msg)
}

// PendingRequestIDs returns stable-sorted pending request IDs.
func (m *SessionManager) PendingRequestIDs() []string {
	if m == nil || m.requests == nil {
		return nil
	}
	return m.requests.PendingRequestIDs()
}

// RequestSnapshot returns one request correlation snapshot.
func (m *SessionManager) RequestSnapshot(requestID string) (RequestSnapshot, bool) {
	if m == nil || m.requests == nil {
		return RequestSnapshot{}, false
	}
	return m.requests.Snapshot(requestID)
}

// Connect transitions closed/reconnecting session to active.
func (m *SessionManager) Connect(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	if m.state == SessionStateClosed || m.state == SessionStateReconnecting {
		m.state = SessionStateConnecting
		m.transportState = TransportStateConnecting
	} else {
		m.mu.Unlock()
		return errors.New("connect invalid in current state")
	}
	m.mu.Unlock()

	if err := m.transport.Connect(ctx, m.session); err != nil {
		m.mu.Lock()
		m.transportState = TransportStateClosed
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			m.state = SessionStateClosed
			m.lastCause = ReconnectCause{Reason: ReconnectReasonConnectFailure, ErrorClass: TransportErrorClassCanceled, Detail: err.Error()}
		} else {
			m.state = SessionStateReconnecting
			m.transportState = TransportStateReconnecting
			m.lastCause = ReconnectCause{Reason: ReconnectReasonConnectFailure, ErrorClass: classifyTransportError(err), Detail: err.Error()}
		}
		m.mu.Unlock()
		return err
	}

	m.mu.Lock()
	m.state = SessionStateActive
	m.transportState = TransportStateConnected
	m.attempt = 0
	m.lastCause = ReconnectCause{Reason: ReconnectReasonNone, ErrorClass: TransportErrorClassUnknown}
	m.mu.Unlock()
	return nil
}

// HandleDisconnect marks the manager as reconnecting when active.
func (m *SessionManager) HandleDisconnect() error {
	return m.HandleDisconnectWithCause(ReconnectCause{Reason: ReconnectReasonDisconnect, ErrorClass: TransportErrorClassTransient})
}

// HandleDisconnectWithCause marks the manager reconnecting with a classified cause.
func (m *SessionManager) HandleDisconnectWithCause(cause ReconnectCause) error {
	if cause.Reason == "" {
		cause.Reason = ReconnectReasonDisconnect
	}
	if cause.ErrorClass == "" {
		cause.ErrorClass = TransportErrorClassUnknown
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.state == SessionStateClosed {
		return errors.New("disconnect invalid in closed state")
	}
	if m.state == SessionStateReconnecting {
		if m.lastCause.Reason == ReconnectReasonNone {
			m.lastCause = cause
		}
		return nil
	}
	m.state = SessionStateReconnecting
	m.transportState = TransportStateReconnecting
	m.lastCause = cause
	return nil
}

// TickHeartbeat updates liveness from a heartbeat protocol message.
func (m *SessionManager) TickHeartbeat(msg RemoteMessage) {
	if isHeartbeat(msg) {
		m.heartbeat.Beat(msg.SentAt)
	}
}

// Reconnect attempts reconnection using retry/backoff policy.
func (m *SessionManager) Reconnect(ctx context.Context) error {
	m.mu.Lock()
	if m.state != SessionStateReconnecting {
		m.mu.Unlock()
		return errors.New("reconnect only valid from reconnecting state")
	}
	if m.reconnecting {
		m.mu.Unlock()
		return errors.New("reconnect already in progress")
	}
	m.reconnecting = true
	maxAttempts := m.policy.MaxAttempts
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		m.reconnecting = false
		m.mu.Unlock()
	}()

	for {
		m.mu.Lock()
		if m.state == SessionStateClosed {
			closeCh := m.closeCh
			m.mu.Unlock()
			select {
			case <-closeCh:
			default:
			}
			return errReconnectAbortedClosed
		}
		if m.state != SessionStateReconnecting {
			m.mu.Unlock()
			return errors.New("reconnect only valid from reconnecting state")
		}
		attempt := m.attempt
		closeCh := m.closeCh
		m.mu.Unlock()

		if maxAttempts > 0 && attempt >= maxAttempts {
			m.mu.Lock()
			m.state = SessionStateClosed
			m.transportState = TransportStateClosed
			m.lastCause = ReconnectCause{
				Reason:         ReconnectReasonReconnectMaxRetries,
				ErrorClass:     TransportErrorClassPermanent,
				ReconnectCount: attempt,
				Detail:         "max reconnect attempts exceeded",
			}
			m.mu.Unlock()
			return errors.New("max reconnect attempts exceeded")
		}

		delay := m.backoff.DelayForAttempt(attempt)
		if delay > 0 {
			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-closeCh:
				timer.Stop()
				return errReconnectAbortedClosed
			case <-timer.C:
			}
		}

		if err := m.transport.Connect(ctx, m.session); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				m.mu.Lock()
				m.lastCause = ReconnectCause{
					Reason:         ReconnectReasonReconnectFailed,
					ErrorClass:     TransportErrorClassCanceled,
					Detail:         err.Error(),
					ReconnectCount: attempt + 1,
				}
				m.mu.Unlock()
				return err
			}
			m.mu.Lock()
			m.attempt++
			m.transportState = TransportStateReconnecting
			m.lastCause = ReconnectCause{
				Reason:         ReconnectReasonReconnectFailed,
				ErrorClass:     classifyTransportError(err),
				Detail:         err.Error(),
				ReconnectCount: m.attempt,
			}
			m.mu.Unlock()
			continue
		}

		m.mu.Lock()
		if m.state == SessionStateClosed {
			m.mu.Unlock()
			_ = m.transport.Close(m.session)
			return errReconnectAbortedClosed
		}
		m.state = SessionStateActive
		m.transportState = TransportStateConnected
		m.attempt = 0
		m.lastCause = ReconnectCause{Reason: ReconnectReasonNone, ErrorClass: TransportErrorClassUnknown}
		m.mu.Unlock()
		return nil
	}
}

// CheckHeartbeat transitions active sessions into reconnecting if stale.
func (m *SessionManager) CheckHeartbeat(now time.Time) bool {
	m.mu.Lock()
	state := m.state
	m.mu.Unlock()
	if state != SessionStateActive {
		return false
	}
	if !m.heartbeat.Expired(now) {
		return false
	}

	m.mu.Lock()
	if m.state == SessionStateActive {
		m.state = SessionStateReconnecting
		m.transportState = TransportStateReconnecting
		m.lastCause = ReconnectCause{Reason: ReconnectReasonHeartbeatExpired, ErrorClass: TransportErrorClassTimeout}
		m.mu.Unlock()
		return true
	}
	m.mu.Unlock()
	return false
}

// Close transitions the session to closed and closes transport.
func (m *SessionManager) Close() error {
	var closeCh chan struct{}
	m.mu.Lock()
	if m.state == SessionStateClosed {
		m.mu.Unlock()
		return nil
	}
	m.state = SessionStateClosed
	m.transportState = TransportStateClosed
	m.lastCause = ReconnectCause{Reason: ReconnectReasonClosed, ErrorClass: TransportErrorClassPermanent}
	closeCh = m.closeCh
	m.closeCh = make(chan struct{})
	m.mu.Unlock()
	close(closeCh)
	err := m.transport.Close(m.session)
	if errors.Is(err, errSessionNotConnected) {
		return nil
	}
	return err
}

func classifyTransportError(err error) TransportErrorClass {
	if err == nil {
		return TransportErrorClassUnknown
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return TransportErrorClassCanceled
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "unauthorized"), strings.Contains(msg, "forbidden"), strings.Contains(msg, "auth"):
		return TransportErrorClassAuth
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "timed out"):
		return TransportErrorClassTimeout
	case strings.Contains(msg, "permanent"), strings.Contains(msg, "invalid session"), strings.Contains(msg, "not found"):
		return TransportErrorClassPermanent
	default:
		return TransportErrorClassTransient
	}
}
