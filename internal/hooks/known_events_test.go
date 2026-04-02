package hooks

import "testing"

func TestIsKnownEvent(t *testing.T) {
	if !isKnownEvent(EventStop) {
		t.Fatalf("expected EventStop to be known")
	}
	if isKnownEvent(Event("custom")) {
		t.Fatalf("did not expect custom event to be known")
	}
}
