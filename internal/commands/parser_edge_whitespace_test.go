package commands

import "testing"

func TestParseCollapsesLeadingAndTrailingWhitespace(t *testing.T) {
	inv, err := Parse("\t /session token show   ")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "session" || len(inv.Args) != 2 {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}
