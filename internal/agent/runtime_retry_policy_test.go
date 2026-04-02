package agent

import (
	"context"
	"errors"
	"testing"
)

func TestDefaultToolRetryPolicy(t *testing.T) {
	policy := defaultToolRetryPolicy()
	if !policy.ShouldRetry(context.Background(), errors.New("temporary timeout")) {
		t.Fatalf("expected retry for timeout")
	}
	if policy.ShouldRetry(context.Background(), context.Canceled) {
		t.Fatalf("expected no retry for canceled")
	}
}

func TestDefaultToolRetryPolicyDeadlineExceeded(t *testing.T) {
	policy := defaultToolRetryPolicy()
	if !policy.ShouldRetry(context.Background(), context.DeadlineExceeded) {
		t.Fatalf("expected retry for deadline exceeded")
	}
}
