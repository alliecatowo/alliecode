package commands

import "testing"

func TestParseMCPDiagnosticsArg(t *testing.T) {
	inv, err := Parse(`/mcp diagnostics`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "mcp" || len(inv.Args) != 1 || inv.Args[0] != "diagnostics" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}
