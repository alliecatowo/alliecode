package bridge

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	defaultPollInterval = 25 * time.Millisecond
	defaultPollBatch    = 16
	defaultPollWait     = 250 * time.Millisecond
)

// RemoteTransport defines remote bridge message delivery behavior.
type RemoteTransport interface {
	Publish(ctx context.Context, env Envelope) error
	Poll(ctx context.Context, receiver string, cfg PollConfig) ([]Envelope, error)
}

// PollConfig tunes remote polling behavior.
type PollConfig struct {
	Interval    time.Duration
	BatchSize   int
	WaitTimeout time.Duration
}

// NormalizePollConfig applies defaults and validation rules.
func NormalizePollConfig(cfg PollConfig) (PollConfig, error) {
	if cfg.Interval <= 0 {
		cfg.Interval = defaultPollInterval
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = defaultPollBatch
	}
	if cfg.WaitTimeout < 0 {
		return PollConfig{}, errors.New("poll wait timeout cannot be negative")
	}
	if cfg.WaitTimeout == 0 {
		cfg.WaitTimeout = defaultPollWait
	}
	return cfg, nil
}

// InMemoryTransport is an in-process transport for tests and local parity.
type InMemoryTransport struct {
	mu     sync.Mutex
	queues map[string][]Envelope
}

// NewInMemoryTransport creates an empty in-memory transport.
func NewInMemoryTransport() *InMemoryTransport {
	return &InMemoryTransport{queues: make(map[string][]Envelope)}
}

// Publish appends an envelope to the receiver queue.
func (t *InMemoryTransport) Publish(ctx context.Context, env Envelope) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := ValidateEnvelope(env); err != nil {
		return fmt.Errorf("validate envelope: %w", err)
	}
	t.mu.Lock()
	t.queues[env.To] = append(t.queues[env.To], env)
	t.mu.Unlock()
	return nil
}

// Poll returns up to BatchSize messages for one receiver.
func (t *InMemoryTransport) Poll(ctx context.Context, receiver string, cfg PollConfig) ([]Envelope, error) {
	if receiver == "" {
		return nil, errors.New("receiver cannot be empty")
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
		queue := t.queues[receiver]
		if len(queue) > 0 {
			n := normalized.BatchSize
			if n > len(queue) {
				n = len(queue)
			}
			out := make([]Envelope, n)
			copy(out, queue[:n])
			remaining := make([]Envelope, len(queue)-n)
			copy(remaining, queue[n:])
			t.queues[receiver] = remaining
			t.mu.Unlock()
			return out, nil
		}
		t.mu.Unlock()

		if normalized.WaitTimeout == 0 || time.Now().After(deadline) {
			return nil, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(normalized.Interval):
		}
	}
}
