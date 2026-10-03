package commands

import "testing"

func TestParseArgsNewAutofixPlan(t *testing.T) {
	inv, err := Parse(`/autofix-pr plan pr-77`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "autofix-pr" {
		t.Fatalf("unexpected name: %q", inv.Name)
	}
	if len(inv.Args) != 2 {
		t.Fatalf("unexpected arg count: %d", len(inv.Args))
	}
	if inv.Args[0] != "plan" {
		t.Fatalf("unexpected arg 0: %q", inv.Args[0])
	}
	if inv.Args[1] != "pr-77" {
		t.Fatalf("unexpected arg 1: %q", inv.Args[1])
	}
}
