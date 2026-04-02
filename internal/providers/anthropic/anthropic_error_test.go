package anthropic

import (
	"net/http"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/providers/common"
)

func TestClassifyAnthropicHTTPError_Auth(t *testing.T) {
	p := &Provider{}
	headers := http.Header{}
	headers.Set("request-id", "req_auth_1")
	err := p.classifyAnthropicHTTPError(http.StatusUnauthorized, headers, []byte(`{"type":"error","error":{"type":"authentication_error","message":"invalid x-api-key"}}`))
	nerr, ok := err.(*common.NormalizedError)
	if !ok {
		t.Fatalf("expected NormalizedError, got %T", err)
	}
	if nerr.Class != common.ErrorClassAuth {
		t.Fatalf("class = %q, want %q", nerr.Class, common.ErrorClassAuth)
	}
	if nerr.Retryable {
		t.Fatalf("expected non-retryable auth error")
	}
	if got := nerr.Message; !containsAll(got, []string{"authentication failed", "request_id=req_auth_1"}) {
		t.Fatalf("message missing diagnostics: %q", got)
	}
}

func TestClassifyAnthropicHTTPError_QuotaNotRetryable(t *testing.T) {
	p := &Provider{}
	err := p.classifyAnthropicHTTPError(http.StatusTooManyRequests, http.Header{}, []byte(`{"type":"error","error":{"type":"rate_limit_error","message":"Your credit balance is too low"}}`))
	nerr, ok := err.(*common.NormalizedError)
	if !ok {
		t.Fatalf("expected NormalizedError, got %T", err)
	}
	if nerr.Class != common.ErrorClassQuota {
		t.Fatalf("class = %q, want %q", nerr.Class, common.ErrorClassQuota)
	}
	if nerr.Retryable {
		t.Fatalf("expected non-retryable quota error")
	}
}

func TestClassifyAnthropicHTTPError_RateLimitHasRetryAfter(t *testing.T) {
	p := &Provider{}
	headers := http.Header{}
	headers.Set("retry-after", "17")
	headers.Set("request-id", "req_rl_1")
	err := p.classifyAnthropicHTTPError(http.StatusTooManyRequests, headers, []byte(`{"type":"error","error":{"type":"rate_limit_error","message":"too many requests"}}`))
	nerr, ok := err.(*common.NormalizedError)
	if !ok {
		t.Fatalf("expected NormalizedError, got %T", err)
	}
	if nerr.Class != common.ErrorClassRateLimit {
		t.Fatalf("class = %q, want %q", nerr.Class, common.ErrorClassRateLimit)
	}
	if !nerr.Retryable {
		t.Fatalf("expected retryable rate-limit error")
	}
	if got := nerr.Message; !containsAll(got, []string{"retry_after=17s", "request_id=req_rl_1"}) {
		t.Fatalf("message missing retry diagnostics: %q", got)
	}
}

func TestClassifyAnthropicHTTPError_ModelNotFound(t *testing.T) {
	p := &Provider{}
	err := p.classifyAnthropicHTTPError(http.StatusNotFound, http.Header{}, []byte(`{"type":"error","error":{"type":"not_found_error","message":"model not found"}}`))
	nerr, ok := err.(*common.NormalizedError)
	if !ok {
		t.Fatalf("expected NormalizedError, got %T", err)
	}
	if nerr.Class != common.ErrorClassModelNotFound {
		t.Fatalf("class = %q, want %q", nerr.Class, common.ErrorClassModelNotFound)
	}
	if nerr.Retryable {
		t.Fatalf("expected model-not-found to be non-retryable")
	}
}

func TestSanitizeDiagnosticRedactsAPIKeys(t *testing.T) {
	input := "invalid key sk-ant-api03-abcdef12345 and sk-test-foo"
	got := sanitizeDiagnostic(input)
	if containsAny(got, []string{"sk-ant-", "sk-test-"}) {
		t.Fatalf("expected key material redacted, got %q", got)
	}
	if !containsAll(got, []string{"[redacted]"}) {
		t.Fatalf("expected redaction marker, got %q", got)
	}
}

func containsAll(s string, parts []string) bool {
	for _, part := range parts {
		if !contains(s, part) {
			return false
		}
	}
	return true
}

func containsAny(s string, parts []string) bool {
	for _, part := range parts {
		if contains(s, part) {
			return true
		}
	}
	return false
}

func contains(s, part string) bool {
	return strings.Contains(s, part)
}
