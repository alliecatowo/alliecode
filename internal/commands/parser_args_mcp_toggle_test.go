package commands

import "testing"

func TestParseMCPToggleArgs(t *testing.T) {
	inv, err := Parse(`/mcp enable all`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "mcp" || len(inv.Args) != 2 || inv.Args[0] != "enable" || inv.Args[1] != "all" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}
