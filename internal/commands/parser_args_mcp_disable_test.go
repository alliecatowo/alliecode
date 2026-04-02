package commands

import "testing"

func TestParseMCPDisableArg(t *testing.T) {
	inv, err := Parse(`/mcp disable local-server`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "mcp" || len(inv.Args) != 2 || inv.Args[0] != "disable" || inv.Args[1] != "local-server" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}
