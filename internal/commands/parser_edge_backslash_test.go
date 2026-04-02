package commands

import "testing"

func TestParseBackslashEscapesNextRuneOutsideQuotes(t *testing.T) {
	inv, err := Parse(`/history list text ticket\ 123`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(inv.Args) != 3 || inv.Args[2] != "ticket 123" {
		t.Fatalf("unexpected args: %#v", inv.Args)
	}
}
