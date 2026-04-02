package common

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// ErrorClass is a normalized user-facing error category.
type ErrorClass string

const (
	ErrorClassAuth          ErrorClass = "auth"
	ErrorClassQuota         ErrorClass = "quota"
	ErrorClassPermission    ErrorClass = "permission"
	ErrorClassRateLimit     ErrorClass = "rate_limit"
	ErrorClassModelNotFound ErrorClass = "model_not_found"
	ErrorClassInvalidInput  ErrorClass = "invalid_input"
	ErrorClassNotFound      ErrorClass = "not_found"
	ErrorClassTimeout       ErrorClass = "timeout"
	ErrorClassTransport     ErrorClass = "transport"
	ErrorClassUnavailable   ErrorClass = "unavailable"
	ErrorClassTransient     ErrorClass = "transient"
	ErrorClassCanceled      ErrorClass = "canceled"
	ErrorClassUnknown       ErrorClass = "unknown"
)

// NormalizedError converts provider-specific failures to consistent classes.
type NormalizedError struct {
	Provider   string
	Class      ErrorClass
	StatusCode int
	Message    string
	Retryable  bool
	Cause      error
}

func (e *NormalizedError) Error() string {
	if e == nil {
		return ""
	}
	if e.Provider == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Provider, e.Message)
}

// Unwrap exposes the original provider error.
func (e *NormalizedError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

// TranslateError normalizes request/network/provider errors.
func TranslateError(provider string, err error) error {
	if err == nil {
		return nil
	}
	if normalized, ok := err.(*NormalizedError); ok {
		return normalized
	}

	if errors.Is(err, context.Canceled) {
		return &NormalizedError{Provider: provider, Class: ErrorClassCanceled, Message: "request canceled", Retryable: false, Cause: err}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return &NormalizedError{Provider: provider, Class: ErrorClassTimeout, Message: "request timed out", Retryable: true, Cause: err}
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return &NormalizedError{Provider: provider, Class: ErrorClassTimeout, Message: "request timed out", Retryable: true, Cause: err}
		}
		return &NormalizedError{Provider: provider, Class: ErrorClassTransport, Message: "transport connection failed", Retryable: true, Cause: err}
	}

	status := parseStatusCode(err.Error())
	if status != 0 {
		return newStatusNormalizedError(provider, status, "", err)
	}

	return &NormalizedError{Provider: provider, Class: ErrorClassUnknown, Message: "provider request failed", Retryable: false, Cause: err}
}

// TranslateHTTPError normalizes non-2xx HTTP responses.
func TranslateHTTPError(provider string, statusCode int, body []byte) error {
	return newStatusNormalizedError(provider, statusCode, string(body), nil)
}

func newStatusNormalizedError(provider string, statusCode int, body string, cause error) *NormalizedError {
	class, retryable := classifyHTTPStatus(statusCode)
	message := strings.TrimSpace(body)
	if message == "" {
		message = defaultStatusMessage(class)
	}
	return &NormalizedError{
		Provider:   provider,
		Class:      class,
		StatusCode: statusCode,
		Message:    message,
		Retryable:  retryable,
		Cause:      cause,
	}
}

func classifyHTTPStatus(statusCode int) (ErrorClass, bool) {
	switch statusCode {
	case 400, 422:
		return ErrorClassInvalidInput, false
	case 401:
		return ErrorClassAuth, false
	case 403:
		return ErrorClassPermission, false
	case 404:
		return ErrorClassNotFound, false
	case 408:
		return ErrorClassTimeout, true
	case 402:
		return ErrorClassQuota, false
	case 409, 425, 429:
		return ErrorClassRateLimit, true
	case 500, 502, 503, 504:
		return ErrorClassUnavailable, true
	default:
		if statusCode >= 500 {
			return ErrorClassUnavailable, true
		}
		if statusCode >= 400 {
			return ErrorClassInvalidInput, false
		}
		return ErrorClassUnknown, false
	}
}

func defaultStatusMessage(class ErrorClass) string {
	switch class {
	case ErrorClassAuth:
		return "authentication failed"
	case ErrorClassQuota:
		return "quota exceeded"
	case ErrorClassPermission:
		return "permission denied"
	case ErrorClassRateLimit:
		return "rate limit exceeded"
	case ErrorClassModelNotFound:
		return "model not found"
	case ErrorClassInvalidInput:
		return "invalid request"
	case ErrorClassNotFound:
		return "resource not found"
	case ErrorClassTimeout:
		return "request timed out"
	case ErrorClassTransport:
		return "transport connection failed"
	case ErrorClassUnavailable:
		return "provider unavailable"
	default:
		return "provider request failed"
	}
}

func parseStatusCode(message string) int {
	idx := strings.Index(message, "API error ")
	if idx < 0 {
		return 0
	}
	rest := message[idx+len("API error "):]
	end := 0
	for end < len(rest) && rest[end] >= '0' && rest[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0
	}
	status, err := strconv.Atoi(rest[:end])
	if err != nil {
		return 0
	}
	return status
}
