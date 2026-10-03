package commands

import "testing"

func TestParseArgsNewBackfillDryrun(t *testing.T) {
	inv, err := Parse(`/backfill-sessions dry-run 3`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "backfill-sessions" {
		t.Fatalf("unexpected name: %q", inv.Name)
	}
	if len(inv.Args) != 2 {
		t.Fatalf("unexpected arg count: %d", len(inv.Args))
	}
	if inv.Args[0] != "dry-run" {
		t.Fatalf("unexpected arg 0: %q", inv.Args[0])
	}
	if inv.Args[1] != "3" {
		t.Fatalf("unexpected arg 1: %q", inv.Args[1])
	}
}
