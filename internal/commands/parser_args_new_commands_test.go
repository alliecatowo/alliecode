package commands

import "testing"

func TestParseArgsAutofixPRPlanRef(t *testing.T) {
	inv, err := Parse(`/autofix-pr plan "pr-123"`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "autofix-pr" || len(inv.Args) != 2 || inv.Args[0] != "plan" || inv.Args[1] != "pr-123" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}

func TestParseArgsBackfillSessionsDryRunCount(t *testing.T) {
	inv, err := Parse(`/backfill-sessions dry-run 25`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "backfill-sessions" || len(inv.Args) != 2 || inv.Args[0] != "dry-run" || inv.Args[1] != "25" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}

func TestParseArgsBreakCacheModels(t *testing.T) {
	inv, err := Parse(`/break-cache models`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "break-cache" || len(inv.Args) != 1 || inv.Args[0] != "models" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}

func TestParseArgsPerfIssueOpenTitle(t *testing.T) {
	inv, err := Parse(`/perf-issue open "slow startup"`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "perf-issue" || len(inv.Args) != 2 || inv.Args[0] != "open" || inv.Args[1] != "slow startup" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}

func TestParseArgsUltraplanRunTarget(t *testing.T) {
	inv, err := Parse(`/ultraplan run release`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if inv.Name != "ultraplan" || len(inv.Args) != 2 || inv.Args[0] != "run" || inv.Args[1] != "release" {
		t.Fatalf("unexpected invocation: %#v", inv)
	}
}
