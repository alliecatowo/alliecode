package commands

import "testing"

func TestParseArgsNewHeapdumpCapture(t *testing.T) {
	inv, err := Parse(`/heapdump capture oom`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "heapdump" {
		t.Fatalf("unexpected name: %q", inv.Name)
	}
	if len(inv.Args) != 2 {
		t.Fatalf("unexpected arg count: %d", len(inv.Args))
	}
	if inv.Args[0] != "capture" {
		t.Fatalf("unexpected arg 0: %q", inv.Args[0])
	}
	if inv.Args[1] != "oom" {
		t.Fatalf("unexpected arg 1: %q", inv.Args[1])
	}
}
