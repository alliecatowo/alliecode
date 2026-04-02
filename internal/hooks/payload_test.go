package hooks

import "testing"

func TestNormalizePayload(t *testing.T) {
	in := map[string]string{"tool-name": "bash", "retry count": "2"}
	out := NormalizePayload(in)
	if out["TOOL_NAME"] != "bash" {
		t.Fatalf("expected TOOL_NAME")
	}
	if out["RETRY_COUNT"] != "2" {
		t.Fatalf("expected RETRY_COUNT")
	}
}
