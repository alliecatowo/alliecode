package commands

import "testing"

func TestParseHistoryCompoundFilterArgs(t *testing.T) {
	inv, err := Parse(`/history list model gpt-4o text "bug fix" limit 2`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "history" || len(inv.Args) != 7 {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}
