package commands

import "testing"

func TestParsePluginRepairArg(t *testing.T) {
	inv, err := Parse(`/plugin repair enable-all`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "plugin" || len(inv.Args) != 2 || inv.Args[1] != "enable-all" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}
