package hooks

import "testing"

func TestClassifyStopReason(t *testing.T) {
	if got := ClassifyStopReason("max_tokens"); got != StopReasonClassModel {
		t.Fatalf("class = %q, want %q", got, StopReasonClassModel)
	}
	if got := ClassifyStopReason("budget_tokens"); got != StopReasonClassBudget {
		t.Fatalf("class = %q, want %q", got, StopReasonClassBudget)
	}
	if got := ClassifyStopReason("hook_error"); got != StopReasonClassHook {
		t.Fatalf("class = %q, want %q", got, StopReasonClassHook)
	}
}

func TestIsTerminalStopReason(t *testing.T) {
	if IsTerminalStopReason("end_turn") {
		t.Fatalf("end_turn should not be terminal")
	}
	if !IsTerminalStopReason("provider_error") {
		t.Fatalf("provider_error should be terminal")
	}
}
