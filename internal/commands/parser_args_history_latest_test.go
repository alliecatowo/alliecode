package commands

import "testing"

func TestParseHistoryLatestListArg(t *testing.T) {
	inv, err := Parse(`/history list latest`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "history" || len(inv.Args) != 2 || inv.Args[1] != "latest" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}
