package commands

import "testing"

func TestParseArgsNewMocklimitsOn(t *testing.T) {
	inv, err := Parse(`/mock-limits on 9`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "mock-limits" {
		t.Fatalf("unexpected name: %q", inv.Name)
	}
	if len(inv.Args) != 2 {
		t.Fatalf("unexpected arg count: %d", len(inv.Args))
	}
	if inv.Args[0] != "on" {
		t.Fatalf("unexpected arg 0: %q", inv.Args[0])
	}
	if inv.Args[1] != "9" {
		t.Fatalf("unexpected arg 1: %q", inv.Args[1])
	}
}
