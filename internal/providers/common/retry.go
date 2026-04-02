package common

import (
	"context"
	"time"
)

// RetryDecision describes whether a failed request should be retried.
type RetryDecision struct {
	Retry      bool
	Backoff    time.Duration
	MaxRetries int
}

// RetryPolicy computes retry behavior for provider errors.
type RetryPolicy interface {
	Decide(err error, attempt int) RetryDecision
}

// DefaultRetryPolicy retries transient/status-based failures with bounded backoff.
type DefaultRetryPolicy struct {
	MaxRetries  int
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
	Table       RetryTable
}

// NewDefaultRetryPolicy returns the default retry policy.
func NewDefaultRetryPolicy() DefaultRetryPolicy {
	return DefaultRetryPolicy{
		MaxRetries:  2,
		BaseBackoff: 250 * time.Millisecond,
		MaxBackoff:  2 * time.Second,
		Table:       DefaultRetryTable(),
	}
}

// Decide returns retry behavior for this attempt and error.
func (p DefaultRetryPolicy) Decide(err error, attempt int) RetryDecision {
	usedDefaultBase := false
	usedDefaultMax := false
	if p.MaxRetries <= 0 {
		p.MaxRetries = 2
	}
	if p.BaseBackoff <= 0 {
		p.BaseBackoff = 250 * time.Millisecond
		usedDefaultBase = true
	}
	if p.MaxBackoff <= 0 {
		p.MaxBackoff = 2 * time.Second
		usedDefaultMax = true
	}
	if p.Table == nil {
		p.Table = DefaultRetryTable()
	}

	if attempt >= p.MaxRetries {
		return RetryDecision{Retry: false, MaxRetries: p.MaxRetries}
	}

	normalized := TranslateError("", err)
	nerr, ok := normalized.(*NormalizedError)
	if !ok || !nerr.Retryable {
		return RetryDecision{Retry: false, MaxRetries: p.MaxRetries}
	}

	profile := p.Table.Lookup(nerr.Class)
	if !profile.Retryable {
		return RetryDecision{Retry: false, MaxRetries: p.MaxRetries}
	}
	maxRetries := p.MaxRetries
	if profile.MaxRetries > 0 {
		maxRetries = profile.MaxRetries
	}
	if attempt >= maxRetries {
		return RetryDecision{Retry: false, MaxRetries: maxRetries}
	}

	multiplier := profile.Multiplier
	if multiplier <= 0 {
		multiplier = 2
	}
	backoff := p.BaseBackoff
	if usedDefaultBase && profile.MinBackoff > 0 {
		backoff = profile.MinBackoff
	}
	for i := 0; i < attempt; i++ {
		backoff *= time.Duration(multiplier)
	}
	maxBackoff := p.MaxBackoff
	if usedDefaultMax && profile.MaxBackoff > 0 {
		maxBackoff = profile.MaxBackoff
	}
	if backoff > maxBackoff {
		backoff = maxBackoff
	}

	return RetryDecision{
		Retry:      true,
		Backoff:    backoff,
		MaxRetries: maxRetries,
	}
}

// WaitBackoff waits for backoff duration or context cancellation.
func WaitBackoff(ctx context.Context, backoff time.Duration) error {
	if backoff <= 0 {
		return nil
	}
	timer := time.NewTimer(backoff)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
