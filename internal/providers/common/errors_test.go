package common

import (
	"context"
	"errors"
	"testing"

	"net"
)

func TestTranslateHTTPErrorClasses(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		class     ErrorClass
		retryable bool
	}{
		{name: "auth", status: 401, class: ErrorClassAuth, retryable: false},
		{name: "rate_limit", status: 429, class: ErrorClassRateLimit, retryable: true},
		{name: "unavailable", status: 503, class: ErrorClassUnavailable, retryable: true},
		{name: "invalid", status: 400, class: ErrorClassInvalidInput, retryable: false},
		{name: "not_found", status: 404, class: ErrorClassNotFound, retryable: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TranslateHTTPError("openai", tt.status, []byte("boom"))
			nerr, ok := err.(*NormalizedError)
			if !ok {
				t.Fatalf("expected NormalizedError, got %T", err)
			}
			if nerr.Class != tt.class {
				t.Fatalf("class = %q, want %q", nerr.Class, tt.class)
			}
			if nerr.Retryable != tt.retryable {
				t.Fatalf("retryable = %v, want %v", nerr.Retryable, tt.retryable)
			}
		})
	}
}

func TestTranslateErrorContextCancellation(t *testing.T) {
	err := TranslateError("anthropic", context.Canceled)
	nerr, ok := err.(*NormalizedError)
	if !ok {
		t.Fatalf("expected NormalizedError, got %T", err)
	}
	if nerr.Class != ErrorClassCanceled {
		t.Fatalf("class = %q, want %q", nerr.Class, ErrorClassCanceled)
	}
}

func TestTranslateErrorFromStatusText(t *testing.T) {
	err := TranslateError("openai", errors.New("openai: API error 429: too many requests"))
	nerr, ok := err.(*NormalizedError)
	if !ok {
		t.Fatalf("expected NormalizedError, got %T", err)
	}
	if nerr.Class != ErrorClassRateLimit {
		t.Fatalf("class = %q, want %q", nerr.Class, ErrorClassRateLimit)
	}
	if !nerr.Retryable {
		t.Fatalf("expected retryable error")
	}
}

func TestTranslateNetworkTimeout(t *testing.T) {
	err := TranslateError("openai", timeoutErr{})
	nerr, ok := err.(*NormalizedError)
	if !ok {
		t.Fatalf("expected NormalizedError, got %T", err)
	}
	if nerr.Class != ErrorClassTimeout {
		t.Fatalf("class = %q, want %q", nerr.Class, ErrorClassTimeout)
	}
}

type timeoutErr struct{}

func (timeoutErr) Error() string   { return "timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return true }

var _ net.Error = timeoutErr{}
