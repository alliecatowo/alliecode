package commands

import "testing"

func TestParseArgsNewRemoteSetupConnect(t *testing.T) {
	inv, err := Parse(`/remote-setup connect`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "remote-setup" {
		t.Fatalf("unexpected name: %q", inv.Name)
	}
	if len(inv.Args) != 1 {
		t.Fatalf("unexpected arg count: %d", len(inv.Args))
	}
	if inv.Args[0] != "connect" {
		t.Fatalf("unexpected arg 0: %q", inv.Args[0])
	}
}
