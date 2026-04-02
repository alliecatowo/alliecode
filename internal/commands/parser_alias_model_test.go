package commands

import "testing"

func TestParseModelAliasCommandName(t *testing.T) {
	inv, err := Parse("/m openai/gpt-4o")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "m" {
		t.Fatalf("expected alias command token m, got %q", inv.Name)
	}
}
