package hooks

import "testing"

func TestBuildPayloadPreservesStopReasonMetadata(t *testing.T) {
	out := BuildPayload(EventStop, map[string]string{
		"stop_reason":          "provider_error",
		"stop_reason_class":    string(StopReasonClassRuntime),
		"stop_reason_terminal": "true",
	})
	if out["STOP_REASON"] != "provider_error" {
		t.Fatalf("missing STOP_REASON in payload")
	}
	if out["STOP_REASON_CLASS"] != string(StopReasonClassRuntime) {
		t.Fatalf("missing STOP_REASON_CLASS in payload")
	}
	if out["STOP_REASON_TERMINAL"] != "true" {
		t.Fatalf("missing STOP_REASON_TERMINAL in payload")
	}
}
