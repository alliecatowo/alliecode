package commands

import "testing"

func TestParseSessionRepairModeArg(t *testing.T) {
	inv, err := Parse(`/session repair reconnect`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "session" || len(inv.Args) != 2 || inv.Args[1] != "reconnect" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}
