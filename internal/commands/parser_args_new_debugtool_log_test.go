package commands

import "testing"

func TestParseArgsNewDebugtoolLog(t *testing.T) {
	inv, err := Parse(`/debug-tool-call log bash`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "debug-tool-call" {
		t.Fatalf("unexpected name: %q", inv.Name)
	}
	if len(inv.Args) != 2 {
		t.Fatalf("unexpected arg count: %d", len(inv.Args))
	}
	if inv.Args[0] != "log" {
		t.Fatalf("unexpected arg 0: %q", inv.Args[0])
	}
	if inv.Args[1] != "bash" {
		t.Fatalf("unexpected arg 1: %q", inv.Args[1])
	}
}
