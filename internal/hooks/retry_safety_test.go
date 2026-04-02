package hooks

import (
	"context"
	"testing"
)

func TestCanRetryHookSafelyTimeoutOnSafeEvent(t *testing.T) {
	if !canRetryHookSafely(EventPreChat, context.DeadlineExceeded) {
		t.Fatalf("expected pre_chat timeout to be retryable")
	}
	if canRetryHookSafely(EventPreTool, context.DeadlineExceeded) {
		t.Fatalf("did not expect pre_tool timeout to be retryable")
	}
}

func TestCanRetryHookSafelyNeverRetriesCanceled(t *testing.T) {
	if canRetryHookSafely(EventPostChat, context.Canceled) {
		t.Fatalf("did not expect canceled error to be retryable")
	}
}
