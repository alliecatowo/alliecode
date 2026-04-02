package commands

import "testing"

func TestParseMCPNoRedirectArg(t *testing.T) {
	inv, err := Parse(`/mcp no-redirect`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "mcp" || len(inv.Args) != 1 || inv.Args[0] != "no-redirect" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}
