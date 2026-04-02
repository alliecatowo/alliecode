package commands

import "testing"

func TestParsePermissionsRetryDenialsArgs(t *testing.T) {
	inv, err := Parse(`/permissions retry-denials`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "permissions" || len(inv.Args) != 1 || inv.Args[0] != "retry-denials" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}
