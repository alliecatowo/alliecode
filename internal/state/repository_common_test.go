package state

import "testing"

func TestMatchesEquals(t *testing.T) {
	if !matchesEquals("CLI", normalizeEquals("cli")) {
		t.Fatalf("expected case-insensitive equality")
	}
	if matchesEquals("cli", normalizeEquals("ide")) {
		t.Fatalf("expected mismatch")
	}
}
