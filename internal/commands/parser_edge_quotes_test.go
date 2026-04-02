package commands

import "testing"

func TestParseSupportsEscapedQuoteInsideQuotedToken(t *testing.T) {
	inv, err := Parse(`/config set note "line \"one\""`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(inv.Args) != 3 || inv.Args[2] != `line "one"` {
		t.Fatalf("unexpected args: %#v", inv.Args)
	}
}
