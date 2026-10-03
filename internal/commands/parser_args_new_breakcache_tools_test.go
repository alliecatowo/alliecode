package commands

import "testing"

func TestParseArgsNewBreakcacheTools(t *testing.T) {
	inv, err := Parse(`/break-cache tools`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "break-cache" {
		t.Fatalf("unexpected name: %q", inv.Name)
	}
	if len(inv.Args) != 1 {
		t.Fatalf("unexpected arg count: %d", len(inv.Args))
	}
	if inv.Args[0] != "tools" {
		t.Fatalf("unexpected arg 0: %q", inv.Args[0])
	}
}
