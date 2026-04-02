package commands

import "testing"

func TestParseConfigRepairProfile(t *testing.T) {
	inv, err := Parse(`/config repair safe`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "config" || len(inv.Args) != 2 || inv.Args[1] != "safe" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}
