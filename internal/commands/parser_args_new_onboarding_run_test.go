package commands

import "testing"

func TestParseArgsNewOnboardingRun(t *testing.T) {
	inv, err := Parse(`/onboarding run`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "onboarding" {
		t.Fatalf("unexpected name: %q", inv.Name)
	}
	if len(inv.Args) != 1 {
		t.Fatalf("unexpected arg count: %d", len(inv.Args))
	}
	if inv.Args[0] != "run" {
		t.Fatalf("unexpected arg 0: %q", inv.Args[0])
	}
}
