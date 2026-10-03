package commands

import "testing"

func TestParseArgsNewUltraplanRun(t *testing.T) {
	inv, err := Parse(`/ultraplan run release`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "ultraplan" {
		t.Fatalf("unexpected name: %q", inv.Name)
	}
	if len(inv.Args) != 2 {
		t.Fatalf("unexpected arg count: %d", len(inv.Args))
	}
	if inv.Args[0] != "run" {
		t.Fatalf("unexpected arg 0: %q", inv.Args[0])
	}
	if inv.Args[1] != "release" {
		t.Fatalf("unexpected arg 1: %q", inv.Args[1])
	}
}
