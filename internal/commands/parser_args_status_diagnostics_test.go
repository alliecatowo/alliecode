package commands

import "testing"

func TestParseStatusDiagnosticsArgument(t *testing.T) {
	inv, err := Parse(`/status diagnostics`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "status" || len(inv.Args) != 1 || inv.Args[0] != "diagnostics" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}
