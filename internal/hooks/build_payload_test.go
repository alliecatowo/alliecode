package hooks

import "testing"

func TestBuildPayloadIncludesStandardFields(t *testing.T) {
	out := BuildPayload(EventRetry, map[string]string{"turn": "2"})
	if out["EVENT"] != string(EventRetry) {
		t.Fatalf("expected EVENT field")
	}
	if out["EVENT_TIME_MS"] == "" || out["SCHEMA"] != "v2" {
		t.Fatalf("expected standard payload fields, got %+v", out)
	}
	if out["TURN"] != "2" {
		t.Fatalf("expected normalized custom field")
	}
}
