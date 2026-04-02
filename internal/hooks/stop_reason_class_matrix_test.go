package hooks

import "testing"

func TestClassifyStopReasonMatrix(t *testing.T) {
	cases := map[string]StopReasonClass{
		"tool_execution_error":          StopReasonClassTool,
		"input_error":                   StopReasonClassInput,
		"max_tokens_recovery_exhausted": StopReasonClassRuntime,
		"unknown-value":                 StopReasonClassUnknown,
	}
	for in, want := range cases {
		if got := ClassifyStopReason(in); got != want {
			t.Fatalf("ClassifyStopReason(%q) = %q, want %q", in, got, want)
		}
	}
}
