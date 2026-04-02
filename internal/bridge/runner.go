package bridge

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"
)

const (
	defaultBackoffInitial = 250 * time.Millisecond
	defaultBackoffMax     = 10 * time.Second
	defaultBackoffFactor  = 2.0
	defaultBackoffJitter  = 0.2
)

// BackoffPolicy controls reconnect retry behavior.
type BackoffPolicy struct {
	Initial time.Duration `json:"initial"`
	Max     time.Duration `json:"max"`
	Factor  float64       `json:"factor"`
	Jitter  float64       `json:"jitter"`
}

// NormalizeBackoffPolicy applies defaults and validates backoff policy.
func NormalizeBackoffPolicy(policy BackoffPolicy) (BackoffPolicy, error) {
	if policy.Initial <= 0 {
		policy.Initial = defaultBackoffInitial
	}
	if policy.Max <= 0 {
		policy.Max = defaultBackoffMax
	}
	if policy.Max < policy.Initial {
		policy.Max = policy.Initial
	}
	if policy.Factor < 1 {
		if policy.Factor == 0 {
			policy.Factor = defaultBackoffFactor
		} else {
			return BackoffPolicy{}, errors.New("backoff factor must be >= 1")
		}
	}
	if policy.Jitter < 0 || policy.Jitter > 1 {
		return BackoffPolicy{}, errors.New("backoff jitter must be between 0 and 1")
	}
	if policy.Jitter == 0 {
		policy.Jitter = defaultBackoffJitter
	}
	return policy, nil
}

// RunnerHooks defines optional lifecycle hooks.
type RunnerHooks struct {
	OnConnected    func()
	OnDisconnected func(error)
	OnHeartbeat    func(time.Time)
}

// RunnerSession defines minimal bridge session operations for runner.
type RunnerSession interface {
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Heartbeat(ctx context.Context) error
}

// SessionRunner keeps one bridge session connected with reconnect/backoff behavior.
type SessionRunner struct {
	Session   RunnerSession
	Backoff   BackoffPolicy
	Heartbeat time.Duration
	Hooks     RunnerHooks

	rand    *rand.Rand
	pending *PendingRequestCorrelator
}

type pendingRequestEntry struct {
	requestID string
	state     RequestLifecycleState
	attempts  int
	reason    string
	cancel    context.CancelFunc
}

// RequestLifecycleState captures pending request terminal transitions.
type RequestLifecycleState string

const (
	RequestLifecycleStatePending   RequestLifecycleState = "pending"
	RequestLifecycleStateResponded RequestLifecycleState = "responded"
	RequestLifecycleStateCanceled  RequestLifecycleState = "canceled"
)

// PendingRequestSnapshot exposes one correlated pending request.
type PendingRequestSnapshot struct {
	RequestID string                `json:"request_id"`
	State     RequestLifecycleState `json:"state"`
	Attempts  int                   `json:"attempts"`
	Reason    string                `json:"reason,omitempty"`
}

// PendingRequestStats aggregates correlation state.
type PendingRequestStats struct {
	Pending       int `json:"pending"`
	Responded     int `json:"responded"`
	Canceled      int `json:"canceled"`
	RetryAttempts int `json:"retry_attempts"`
}

// PendingRequestCorrelator tracks pending requests and resolves cancel/response lifecycle events.
type PendingRequestCorrelator struct {
	mu      sync.Mutex
	pending map[string]pendingRequestEntry
	stats   PendingRequestStats
}

// NewPendingRequestCorrelator creates an empty pending request correlator.
func NewPendingRequestCorrelator() *PendingRequestCorrelator {
	return &PendingRequestCorrelator{pending: make(map[string]pendingRequestEntry)}
}

// Register stores one request ID and optional cancellation callback.
func (c *PendingRequestCorrelator) Register(requestID string, cancel context.CancelFunc) error {
	if c == nil {
		return errors.New("pending request correlator cannot be nil")
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return errors.New("request ID cannot be empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if existing, exists := c.pending[requestID]; exists {
		if existing.state != RequestLifecycleStatePending {
			return errors.New("pending request already resolved")
		}
		existing.attempts++
		if cancel != nil {
			existing.cancel = cancel
		}
		c.pending[requestID] = existing
		c.stats.RetryAttempts++
		return nil
	}
	c.pending[requestID] = pendingRequestEntry{
		requestID: requestID,
		state:     RequestLifecycleStatePending,
		attempts:  1,
		cancel:    cancel,
	}
	c.stats.Pending++
	return nil
}

// ResolveResponse removes one pending request for an inbound response.
func (c *PendingRequestCorrelator) ResolveResponse(resp ResponseWrapper) bool {
	if c == nil {
		return false
	}
	requestID := strings.TrimSpace(resp.RequestID)
	if requestID == "" {
		return false
	}

	c.mu.Lock()
	entry, ok := c.pending[requestID]
	if ok {
		entry.state = RequestLifecycleStateResponded
		entry.reason = ""
		c.pending[requestID] = entry
		c.stats.Pending--
		c.stats.Responded++
		delete(c.pending, requestID)
	}
	c.mu.Unlock()
	return ok
}

// ResolveCancel removes one pending request and triggers its cancellation callback.
func (c *PendingRequestCorrelator) ResolveCancel(cancel CancelWrapper) bool {
	if c == nil {
		return false
	}
	requestID := strings.TrimSpace(cancel.RequestID)
	if requestID == "" {
		return false
	}

	c.mu.Lock()
	entry, ok := c.pending[requestID]
	if ok {
		entry.state = RequestLifecycleStateCanceled
		entry.reason = cancel.Reason
		c.pending[requestID] = entry
		c.stats.Pending--
		c.stats.Canceled++
		delete(c.pending, requestID)
	}
	c.mu.Unlock()
	if ok && entry.cancel != nil {
		entry.cancel()
	}
	return ok
}

// CancelAll resolves all pending requests through their cancellation path.
func (c *PendingRequestCorrelator) CancelAll() {
	if c == nil {
		return
	}

	c.mu.Lock()
	entries := make([]pendingRequestEntry, 0, len(c.pending))
	for _, entry := range c.pending {
		entry.state = RequestLifecycleStateCanceled
		entry.reason = "runner_disconnected"
		entries = append(entries, entry)
		c.stats.Pending--
		c.stats.Canceled++
	}
	c.pending = make(map[string]pendingRequestEntry)
	c.mu.Unlock()

	for _, entry := range entries {
		if entry.cancel != nil {
			entry.cancel()
		}
	}
}

// PendingCount returns the number of currently tracked pending requests.
func (c *PendingRequestCorrelator) PendingCount() int {
	if c == nil {
		return 0
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.pending)
}

// Stats reports aggregate pending request correlation state.
func (c *PendingRequestCorrelator) Stats() PendingRequestStats {
	if c == nil {
		return PendingRequestStats{}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.stats
}

// Snapshot returns one pending request snapshot when present.
func (c *PendingRequestCorrelator) Snapshot(requestID string) (PendingRequestSnapshot, bool) {
	if c == nil {
		return PendingRequestSnapshot{}, false
	}
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return PendingRequestSnapshot{}, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.pending[requestID]
	if !ok {
		return PendingRequestSnapshot{}, false
	}
	return PendingRequestSnapshot{RequestID: entry.requestID, State: entry.state, Attempts: entry.attempts, Reason: entry.reason}, true
}

// NewSessionRunner creates a session runner with defaults.
func NewSessionRunner(session RunnerSession) (*SessionRunner, error) {
	if session == nil {
		return nil, errors.New("runner session cannot be nil")
	}
	backoff, err := NormalizeBackoffPolicy(BackoffPolicy{})
	if err != nil {
		return nil, err
	}
	return &SessionRunner{
		Session:   session,
		Backoff:   backoff,
		Heartbeat: defaultHeartbeatInterval,
		rand:      rand.New(rand.NewSource(time.Now().UnixNano())),
		pending:   NewPendingRequestCorrelator(),
	}, nil
}

// RegisterPendingRequest tracks one outbound request for response/cancel correlation.
func (r *SessionRunner) RegisterPendingRequest(requestID string, cancel context.CancelFunc) error {
	if r == nil || r.pending == nil {
		return errors.New("session runner pending correlator is not initialized")
	}
	return r.pending.Register(requestID, cancel)
}

// ResolvePendingResponse clears one pending request when a response arrives.
func (r *SessionRunner) ResolvePendingResponse(resp ResponseWrapper) bool {
	if r == nil || r.pending == nil {
		return false
	}
	return r.pending.ResolveResponse(resp)
}

// ResolvePendingCancel clears one pending request through the cancel path.
func (r *SessionRunner) ResolvePendingCancel(cancel CancelWrapper) bool {
	if r == nil || r.pending == nil {
		return false
	}
	return r.pending.ResolveCancel(cancel)
}

// Run starts the session loop until context cancellation.
func (r *SessionRunner) Run(ctx context.Context) error {
	if r.Heartbeat <= 0 {
		r.Heartbeat = defaultHeartbeatInterval
	}
	policy, err := NormalizeBackoffPolicy(r.Backoff)
	if err != nil {
		return err
	}
	r.Backoff = policy

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := r.Session.Connect(ctx); err != nil {
			if !r.waitBackoff(ctx, 0) {
				return ctx.Err()
			}
			continue
		}
		if r.Hooks.OnConnected != nil {
			r.Hooks.OnConnected()
		}

		attempt := 0
		err = r.runConnectedLoop(ctx)
		if r.Hooks.OnDisconnected != nil {
			r.Hooks.OnDisconnected(err)
		}
		r.pending.CancelAll()
		_ = r.Session.Disconnect(context.Background())
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return err
		}
		attempt++
		if !r.waitBackoff(ctx, attempt) {
			return ctx.Err()
		}
	}
}

func (r *SessionRunner) runConnectedLoop(ctx context.Context) error {
	ticker := time.NewTicker(r.Heartbeat)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case at := <-ticker.C:
			if err := r.Session.Heartbeat(ctx); err != nil {
				return err
			}
			if r.Hooks.OnHeartbeat != nil {
				r.Hooks.OnHeartbeat(at.UTC())
			}
		}
	}
}

func (r *SessionRunner) waitBackoff(ctx context.Context, attempt int) bool {
	delay := r.backoffDelay(attempt)
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (r *SessionRunner) backoffDelay(attempt int) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}
	base := float64(r.Backoff.Initial) * math.Pow(r.Backoff.Factor, float64(attempt-1))
	if base > float64(r.Backoff.Max) {
		base = float64(r.Backoff.Max)
	}
	jitterScale := 1.0
	if r.Backoff.Jitter > 0 {
		jitterScale = 1 + ((r.rand.Float64()*2 - 1) * r.Backoff.Jitter)
	}
	out := time.Duration(base * jitterScale)
	if out < r.Backoff.Initial {
		out = r.Backoff.Initial
	}
	if out > r.Backoff.Max {
		out = r.Backoff.Max
	}
	return out
}
