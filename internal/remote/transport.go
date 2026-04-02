package remote

import (
	"context"
	"errors"
	"sync"
	"time"
)

var errSessionNotConnected = errors.New("session not connected")

const (
	defaultPollInterval = 20 * time.Millisecond
	defaultPollBatch    = 32
	defaultPollWait     = 200 * time.Millisecond
)

// Transport describes websocket/sse-like message stream behavior.
type Transport interface {
	Connect(ctx context.Context, sessionID string) error
	Send(ctx context.Context, msg RemoteMessage) error
	Receive(ctx context.Context, sessionID string, cfg PollConfig) ([]RemoteMessage, error)
	Close(sessionID string) error
}

// PollConfig configures polling for receive calls.
type PollConfig struct {
	Interval    time.Duration
	BatchSize   int
	WaitTimeout time.Duration
}

// NormalizePollConfig validates and applies defaults.
func NormalizePollConfig(cfg PollConfig) (PollConfig, error) {
	if cfg.Interval <= 0 {
		cfg.Interval = defaultPollInterval
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaultPollBatch
	}
	if cfg.WaitTimeout < 0 {
		return PollConfig{}, errors.New("wait timeout cannot be negative")
	}
	if cfg.WaitTimeout == 0 {
		cfg.WaitTimeout = defaultPollWait
	}
	return cfg, nil
}

// InMemoryTransport is an in-process transport for tests and local usage.
type InMemoryTransport struct {
	mu      sync.Mutex
	open    map[string]bool
	streams map[string][]RemoteMessage
}

// NewInMemoryTransport returns an empty in-memory transport.
func NewInMemoryTransport() *InMemoryTransport {
	return &InMemoryTransport{
		open:    make(map[string]bool),
		streams: make(map[string][]RemoteMessage),
	}
}

// Connect marks a session stream as active.
func (t *InMemoryTransport) Connect(ctx context.Context, sessionID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if sessionID == "" {
		return errors.New("sessionID cannot be empty")
	}
	t.mu.Lock()
	t.open[sessionID] = true
	t.mu.Unlock()
	return nil
}

// Send appends a message into one session stream.
func (t *InMemoryTransport) Send(ctx context.Context, msg RemoteMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := msg.Validate(); err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.open[msg.SessionID] {
		return errSessionNotConnected
	}
	t.streams[msg.SessionID] = append(t.streams[msg.SessionID], msg)
	return nil
}

// Receive drains up to BatchSize messages for one session.
func (t *InMemoryTransport) Receive(ctx context.Context, sessionID string, cfg PollConfig) ([]RemoteMessage, error) {
	if sessionID == "" {
		return nil, errors.New("sessionID cannot be empty")
	}
	normalized, err := NormalizePollConfig(cfg)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(normalized.WaitTimeout)

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		t.mu.Lock()
		if !t.open[sessionID] {
			t.mu.Unlock()
			return nil, errSessionNotConnected
		}
		stream := t.streams[sessionID]
		if len(stream) > 0 {
			n := normalized.BatchSize
			if n > len(stream) {
				n = len(stream)
			}
			out := make([]RemoteMessage, n)
			copy(out, stream[:n])
			remaining := make([]RemoteMessage, len(stream)-n)
			copy(remaining, stream[n:])
			t.streams[sessionID] = remaining
			t.mu.Unlock()
			return out, nil
		}
		t.mu.Unlock()

		if time.Now().After(deadline) {
			return nil, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(normalized.Interval):
		}
	}
}

// Close marks a session stream as closed.
func (t *InMemoryTransport) Close(sessionID string) error {
	if sessionID == "" {
		return errors.New("sessionID cannot be empty")
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if !t.open[sessionID] {
		return errSessionNotConnected
	}
	delete(t.open, sessionID)
	delete(t.streams, sessionID)
	return nil
}
