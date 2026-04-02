package commands

import "testing"

func TestParseConfigPanelArgs(t *testing.T) {
	inv, err := Parse(`/config panel`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "config" || len(inv.Args) != 1 || inv.Args[0] != "panel" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}
