package remote

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// RetryPolicy controls reconnect delay behavior.
type RetryPolicy struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       float64
	MaxAttempts  int
}

// DefaultRetryPolicy returns conservative reconnect defaults.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     2 * time.Second,
		Multiplier:   2,
		Jitter:       0.15,
		MaxAttempts:  8,
	}
}

// Backoff computes reconnect delays from a retry policy.
type Backoff struct {
	policy RetryPolicy
	rnd    *rand.Rand
	mu     sync.Mutex
}

// NewBackoff builds a backoff calculator.
func NewBackoff(policy RetryPolicy, seed int64) *Backoff {
	if policy.InitialDelay <= 0 {
		policy.InitialDelay = 100 * time.Millisecond
	}
	if policy.MaxDelay <= 0 {
		policy.MaxDelay = 2 * time.Second
	}
	if policy.Multiplier < 1 {
		policy.Multiplier = 1
	}
	if policy.Jitter < 0 {
		policy.Jitter = 0
	}
	if policy.Jitter > 1 {
		policy.Jitter = 1
	}

	return &Backoff{policy: policy, rnd: rand.New(rand.NewSource(seed))}
}

// DelayForAttempt returns wait duration for one attempt (0-indexed).
func (b *Backoff) DelayForAttempt(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	base := float64(b.policy.InitialDelay)
	d := base * math.Pow(b.policy.Multiplier, float64(attempt))
	if d > float64(b.policy.MaxDelay) {
		d = float64(b.policy.MaxDelay)
	}

	b.mu.Lock()
	defer b.mu.Unlock()
	if b.policy.Jitter == 0 {
		return time.Duration(d)
	}

	span := d * b.policy.Jitter
	offset := (b.rnd.Float64()*2 - 1) * span
	result := d + offset
	if result < 0 {
		result = 0
	}
	return time.Duration(result)
}

// HeartbeatMonitor tracks liveness from heartbeat messages.
type HeartbeatMonitor struct {
	timeout  time.Duration
	lastBeat time.Time
	mu       sync.Mutex
}

// NewHeartbeatMonitor creates a monitor with a timeout.
func NewHeartbeatMonitor(timeout time.Duration, now time.Time) *HeartbeatMonitor {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &HeartbeatMonitor{timeout: timeout, lastBeat: now}
}

// Beat records the latest heartbeat timestamp.
func (h *HeartbeatMonitor) Beat(ts time.Time) {
	h.mu.Lock()
	h.lastBeat = ts
	h.mu.Unlock()
}

// Expired reports whether heartbeat timeout has elapsed.
func (h *HeartbeatMonitor) Expired(now time.Time) bool {
	h.mu.Lock()
	last := h.lastBeat
	timeout := h.timeout
	h.mu.Unlock()
	if last.IsZero() {
		return false
	}
	return now.Sub(last) > timeout
}

// LastBeat returns the most recent heartbeat timestamp.
func (h *HeartbeatMonitor) LastBeat() time.Time {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.lastBeat
}
