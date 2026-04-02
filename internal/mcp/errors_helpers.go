package mcp

import (
	"context"
	"errors"
	"strings"
)

type ErrorCategory string

const (
	ErrorCategoryUnknown              ErrorCategory = "unknown"
	ErrorCategoryNeedsAuthentication  ErrorCategory = "needs_authentication"
	ErrorCategoryAuthenticationFailed ErrorCategory = "authentication_failed"
	ErrorCategoryTransportUnsupported ErrorCategory = "transport_unsupported"
	ErrorCategoryConfiguration        ErrorCategory = "configuration"
	ErrorCategoryConnection           ErrorCategory = "connection"
	ErrorCategoryTimeout              ErrorCategory = "timeout"
	ErrorCategoryUnavailable          ErrorCategory = "unavailable"
	ErrorCategoryRateLimited          ErrorCategory = "rate_limited"
	ErrorCategoryInvalidRequest       ErrorCategory = "invalid_request"
	ErrorCategoryExecution            ErrorCategory = "execution"
)

type ClassifiedError struct {
	Category  ErrorCategory `json:"category"`
	Code      string        `json:"code,omitempty"`
	Message   string        `json:"message"`
	Hint      string        `json:"hint,omitempty"`
	Retryable bool          `json:"retryable,omitempty"`
}

func cleanMCPErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return strings.TrimSpace(err.Error())
}

func ClassifyError(err error) ClassifiedError {
	if err == nil {
		return ClassifiedError{Category: ErrorCategoryUnknown, Code: "mcp_error_unknown", Message: ""}
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ClassifiedError{Category: ErrorCategoryTimeout, Code: "mcp_error_timeout", Message: cleanMCPErrorMessage(err), Hint: "Increase timeout or retry after verifying server responsiveness", Retryable: true}
	}
	msg := strings.ToLower(cleanMCPErrorMessage(err))
	if errors.Is(err, ErrTransportUnsupported) {
		return ClassifiedError{Category: ErrorCategoryTransportUnsupported, Code: "mcp_error_transport_unsupported", Message: cleanMCPErrorMessage(err), Hint: "Use a server transport that supports this operation"}
	}
	if strings.Contains(msg, "needs auth") || strings.Contains(msg, "reauth") || strings.Contains(msg, "login required") {
		return ClassifiedError{Category: ErrorCategoryNeedsAuthentication, Code: "mcp_error_needs_authentication", Message: cleanMCPErrorMessage(err), Hint: "Run authentication flow for the MCP server"}
	}
	if strings.Contains(msg, "unauthorized") || strings.Contains(msg, "authentication") || strings.Contains(msg, "401") || strings.Contains(msg, "forbidden") {
		return ClassifiedError{Category: ErrorCategoryAuthenticationFailed, Code: "mcp_error_authentication_failed", Message: cleanMCPErrorMessage(err), Hint: "Verify credentials/token for the MCP server"}
	}
	if strings.Contains(msg, "unknown mcp server") || strings.Contains(msg, "server name cannot be empty") || strings.Contains(msg, "already exists") {
		return ClassifiedError{Category: ErrorCategoryConfiguration, Code: "mcp_error_configuration", Message: cleanMCPErrorMessage(err), Hint: "Check MCP server configuration values"}
	}
	if strings.Contains(msg, "too many requests") || strings.Contains(msg, "429") {
		return ClassifiedError{Category: ErrorCategoryRateLimited, Code: "mcp_error_rate_limited", Message: cleanMCPErrorMessage(err), Hint: "Retry with backoff", Retryable: true}
	}
	if strings.Contains(msg, "bad request") || strings.Contains(msg, "invalid args") || strings.Contains(msg, "invalid input") {
		return ClassifiedError{Category: ErrorCategoryInvalidRequest, Code: "mcp_error_invalid_request", Message: cleanMCPErrorMessage(err), Hint: "Validate tool arguments and request payload"}
	}
	if strings.Contains(msg, "service unavailable") || strings.Contains(msg, "temporarily unavailable") || strings.Contains(msg, "503") {
		return ClassifiedError{Category: ErrorCategoryUnavailable, Code: "mcp_error_unavailable", Message: cleanMCPErrorMessage(err), Hint: "Retry when server is available", Retryable: true}
	}
	if strings.Contains(msg, "initialize") || strings.Contains(msg, "connect") || strings.Contains(msg, "closed") || strings.Contains(msg, "eof") {
		return ClassifiedError{Category: ErrorCategoryConnection, Code: "mcp_error_connection", Message: cleanMCPErrorMessage(err), Hint: "Check server process/network connectivity", Retryable: true}
	}
	if strings.Contains(msg, "failed") || strings.Contains(msg, "panic") || strings.Contains(msg, "exception") {
		return ClassifiedError{Category: ErrorCategoryExecution, Code: "mcp_error_execution", Message: cleanMCPErrorMessage(err)}
	}
	return ClassifiedError{Category: ErrorCategoryUnknown, Code: "mcp_error_unknown", Message: cleanMCPErrorMessage(err)}
}
