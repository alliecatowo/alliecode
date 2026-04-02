package hooks

import (
	"strings"
	"testing"
)

func TestNormalizePayloadTrimsLargeValues(t *testing.T) {
	value := strings.Repeat("x", maxPayloadValueLength+50)
	out := NormalizePayload(map[string]string{"big": value})
	if len(out["BIG"]) != maxPayloadValueLength {
		t.Fatalf("value length = %d, want %d", len(out["BIG"]), maxPayloadValueLength)
	}
}
