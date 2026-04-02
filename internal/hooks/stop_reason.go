package hooks

import "strings"

// StopReasonClass groups stop reasons into stable hook-friendly families.
type StopReasonClass string

const (
	StopReasonClassModel   StopReasonClass = "model"
	StopReasonClassBudget  StopReasonClass = "budget"
	StopReasonClassRuntime StopReasonClass = "runtime"
	StopReasonClassTool    StopReasonClass = "tool"
	StopReasonClassHook    StopReasonClass = "hook"
	StopReasonClassInput   StopReasonClass = "input"
	StopReasonClassUnknown StopReasonClass = "unknown"
)

// ClassifyStopReason returns a stable class for a stop-reason identifier.
func ClassifyStopReason(reason string) StopReasonClass {
	reason = strings.TrimSpace(reason)
	switch reason {
	case "end_turn", "max_tokens", "context_window_exceeded", "stop_sequence", "content_filter", "refusal":
		return StopReasonClassModel
	case "budget_usd", "budget_tokens":
		return StopReasonClassBudget
	case "context_canceled", "provider_error", "persistence_error", "max_turns", "max_tokens_recovery_exhausted":
		return StopReasonClassRuntime
	case "tool_execution_error", "tool_use_malformed":
		return StopReasonClassTool
	case "hook_error":
		return StopReasonClassHook
	case "input_error":
		return StopReasonClassInput
	default:
		return StopReasonClassUnknown
	}
}

// IsTerminalStopReason reports if stop reason should be treated as final terminal state.
func IsTerminalStopReason(reason string) bool {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return false
	}
	switch reason {
	case "end_turn":
		return false
	default:
		return true
	}
}
