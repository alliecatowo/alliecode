package bridge

import (
	"context"
	"errors"
	"sync"
	"time"
)

// PermissionAsk captures one remote permission request.
type PermissionAsk struct {
	SessionID   string
	DeviceID    string
	Permission  string
	Reason      string
	RequestedAt time.Time
}

// PermissionDecision is the result of one permission request.
type PermissionDecision struct {
	Allowed bool
	Message string
	TTL     time.Duration
}

// PermissionCallback handles remote permission asks.
type PermissionCallback interface {
	AskPermission(ctx context.Context, ask PermissionAsk) (PermissionDecision, error)
}

// PermissionCallbackFunc adapts a function to PermissionCallback.
type PermissionCallbackFunc func(ctx context.Context, ask PermissionAsk) (PermissionDecision, error)

// AskPermission invokes wrapped function.
func (f PermissionCallbackFunc) AskPermission(ctx context.Context, ask PermissionAsk) (PermissionDecision, error) {
	if f == nil {
		return PermissionDecision{}, errors.New("permission callback is nil")
	}
	return f(ctx, ask)
}

// StaticPermissionCallback returns the same decision for every ask.
type StaticPermissionCallback struct {
	Decision PermissionDecision
}

// AskPermission returns static decision.
func (s StaticPermissionCallback) AskPermission(context.Context, PermissionAsk) (PermissionDecision, error) {
	return s.Decision, nil
}

// PermissionBroker dispatches permission asks to a callback.
type PermissionBroker struct {
	mu       sync.RWMutex
	callback PermissionCallback
}

// NewPermissionBroker constructs a permission broker.
func NewPermissionBroker(callback PermissionCallback) *PermissionBroker {
	return &PermissionBroker{callback: callback}
}

// SetCallback swaps callback implementation.
func (b *PermissionBroker) SetCallback(callback PermissionCallback) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.callback = callback
}

// Ask executes one permission request.
func (b *PermissionBroker) Ask(ctx context.Context, ask PermissionAsk) (PermissionDecision, error) {
	if _, err := NormalizeSessionID(ask.SessionID); err != nil {
		return PermissionDecision{}, err
	}
	if ask.Permission == "" {
		return PermissionDecision{}, errors.New("permission cannot be empty")
	}
	b.mu.RLock()
	callback := b.callback
	b.mu.RUnlock()
	if callback == nil {
		return PermissionDecision{}, errors.New("permission callback is not configured")
	}
	if ask.RequestedAt.IsZero() {
		ask.RequestedAt = time.Now().UTC()
	}
	return callback.AskPermission(ctx, ask)
}
