package commands

import "testing"

func TestParseSingleQuotedTokenSupportsEscapes(t *testing.T) {
	inv, err := Parse(`/skills add 'lint\'flow'`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(inv.Args) != 2 || inv.Args[1] != "lint'flow" {
		t.Fatalf("unexpected args: %#v", inv.Args)
	}
}
