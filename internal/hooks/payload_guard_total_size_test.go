package hooks

import (
	"strings"
	"testing"
)

func TestBuildPayloadGuardedTruncatesTotalPayload(t *testing.T) {
	value := strings.Repeat("x", 200)
	out := BuildPayloadGuarded(EventPreChat, map[string]string{"a": value, "b": value, "c": value}, PayloadGuard{MaxTotalBytes: 160, MaxValueLength: 200})
	if out["PAYLOAD_TRUNCATED"] != "1" {
		t.Fatalf("expected payload truncation marker")
	}
}

func TestBuildPayloadGuardedTruncatesValueLength(t *testing.T) {
	out := BuildPayloadGuarded(EventPreChat, map[string]string{"k": strings.Repeat("y", 50)}, PayloadGuard{MaxTotalBytes: 4096, MaxValueLength: 10})
	if len(out["K"]) != 10 {
		t.Fatalf("value length = %d, want 10", len(out["K"]))
	}
}
