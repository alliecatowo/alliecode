package common

import (
	"context"
	"testing"
	"time"
)

func TestDefaultRetryPolicyByErrorClass(t *testing.T) {
	policy := DefaultRetryPolicy{MaxRetries: 3, BaseBackoff: time.Millisecond, MaxBackoff: 10 * time.Millisecond}

	retryable := &NormalizedError{Class: ErrorClassRateLimit, Retryable: true}
	decision := policy.Decide(retryable, 0)
	if !decision.Retry {
		t.Fatalf("expected retry on rate limit")
	}
	if decision.Backoff != time.Millisecond {
		t.Fatalf("backoff = %s, want %s", decision.Backoff, time.Millisecond)
	}

	nonRetryable := &NormalizedError{Class: ErrorClassAuth, Retryable: false}
	decision = policy.Decide(nonRetryable, 0)
	if decision.Retry {
		t.Fatalf("expected no retry on auth errors")
	}

	quota := &NormalizedError{Class: ErrorClassQuota, Retryable: true}
	decision = policy.Decide(quota, 0)
	if decision.Retry {
		t.Fatalf("expected no retry on quota errors")
	}

	modelMissing := &NormalizedError{Class: ErrorClassModelNotFound, Retryable: true}
	decision = policy.Decide(modelMissing, 0)
	if decision.Retry {
		t.Fatalf("expected no retry on model-not-found errors")
	}
}

func TestDefaultRetryPolicyStopsAtMaxRetries(t *testing.T) {
	policy := DefaultRetryPolicy{MaxRetries: 2, BaseBackoff: time.Millisecond, MaxBackoff: 10 * time.Millisecond}
	err := &NormalizedError{Class: ErrorClassUnavailable, Retryable: true}

	decision := policy.Decide(err, 2)
	if decision.Retry {
		t.Fatalf("expected retries exhausted")
	}
}

func TestWaitBackoffContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := WaitBackoff(ctx, 25*time.Millisecond)
	if err == nil {
		t.Fatalf("expected context cancellation")
	}
}
