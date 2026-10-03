package commands

import "testing"

func TestParseArgsNewGoodclaudeAsk(t *testing.T) {
	inv, err := Parse(`/good-claude ask "why this test"`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "good-claude" {
		t.Fatalf("unexpected name: %q", inv.Name)
	}
	if len(inv.Args) != 2 {
		t.Fatalf("unexpected arg count: %d", len(inv.Args))
	}
	if inv.Args[0] != "ask" {
		t.Fatalf("unexpected arg 0: %q", inv.Args[0])
	}
	if inv.Args[1] != "why this test" {
		t.Fatalf("unexpected arg 1: %q", inv.Args[1])
	}
}
