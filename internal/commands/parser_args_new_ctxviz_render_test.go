package commands

import "testing"

func TestParseArgsNewCtxvizRender(t *testing.T) {
	inv, err := Parse(`/ctx-viz render timeline`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "ctx-viz" {
		t.Fatalf("unexpected name: %q", inv.Name)
	}
	if len(inv.Args) != 2 {
		t.Fatalf("unexpected arg count: %d", len(inv.Args))
	}
	if inv.Args[0] != "render" {
		t.Fatalf("unexpected arg 0: %q", inv.Args[0])
	}
	if inv.Args[1] != "timeline" {
		t.Fatalf("unexpected arg 1: %q", inv.Args[1])
	}
}
