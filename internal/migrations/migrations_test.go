package migrations

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestRunnerRunAll_RunOnceBehavior(t *testing.T) {
	ledger := NewLedger(filepath.Join(t.TempDir(), "migrations.jsonl"))
	runner := NewRunner(ledger)

	count := 0
	migs := []Migration{
		{Version: "20260331_001", Description: "seed defaults", Run: func() error { count++; return nil }},
		{Version: "20260331_002", Description: "add denials ledger", Run: func() error { count++; return nil }},
	}

	applied, err := runner.RunAll(migs)
	if err != nil {
		t.Fatalf("first RunAll failed: %v", err)
	}
	if got, want := len(applied), 2; got != want {
		t.Fatalf("first RunAll applied %d migrations, want %d", got, want)
	}

	applied, err = runner.RunAll(migs)
	if err != nil {
		t.Fatalf("second RunAll failed: %v", err)
	}
	if got := len(applied); got != 0 {
		t.Fatalf("second RunAll should apply nothing, got %d", got)
	}
	if got, want := count, 2; got != want {
		t.Fatalf("Run functions executed %d times, want %d", got, want)
	}
}

func TestRunnerRunAll_StopsAndDoesNotRecordOnFailure(t *testing.T) {
	ledger := NewLedger(filepath.Join(t.TempDir(), "migrations.jsonl"))
	runner := NewRunner(ledger)

	migs := []Migration{
		{Version: "20260331_001", Run: func() error { return nil }},
		{Version: "20260331_002", Run: func() error { return fmt.Errorf("boom") }},
	}

	_, err := runner.RunAll(migs)
	if err == nil {
		t.Fatalf("expected migration failure")
	}

	entries, err := ledger.Entries()
	if err != nil {
		t.Fatalf("Entries failed: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 recorded migration, got %d", len(entries))
	}
	if entries[0].Version != "20260331_001" {
		t.Fatalf("recorded wrong migration: %q", entries[0].Version)
	}
}

func TestRunnerRunAll_RejectsDuplicateVersions(t *testing.T) {
	ledger := NewLedger(filepath.Join(t.TempDir(), "migrations.jsonl"))
	runner := NewRunner(ledger)

	_, err := runner.RunAll([]Migration{
		{Version: "1", Run: func() error { return nil }},
		{Version: "1", Run: func() error { return nil }},
	})
	if err == nil {
		t.Fatalf("expected duplicate version error")
	}
}

func TestRunnerRunAll_DeterministicVersionOrder(t *testing.T) {
	ledger := NewLedger(filepath.Join(t.TempDir(), "migrations.jsonl"))
	runner := NewRunner(ledger)

	executed := make([]string, 0, 3)
	migs := []Migration{
		{Version: "20260331_003", Run: func() error { executed = append(executed, "20260331_003"); return nil }},
		{Version: "20260331_001", Run: func() error { executed = append(executed, "20260331_001"); return nil }},
		{Version: "20260331_002", Run: func() error { executed = append(executed, "20260331_002"); return nil }},
	}

	applied, err := runner.RunAll(migs)
	if err != nil {
		t.Fatalf("RunAll failed: %v", err)
	}

	want := []string{"20260331_001", "20260331_002", "20260331_003"}
	if len(applied) != len(want) {
		t.Fatalf("applied length = %d, want %d", len(applied), len(want))
	}
	for i := range want {
		if applied[i] != want[i] {
			t.Fatalf("applied[%d] = %q, want %q", i, applied[i], want[i])
		}
		if executed[i] != want[i] {
			t.Fatalf("executed[%d] = %q, want %q", i, executed[i], want[i])
		}
	}

	entries, err := ledger.Entries()
	if err != nil {
		t.Fatalf("Entries failed: %v", err)
	}
	if len(entries) != len(want) {
		t.Fatalf("entries length = %d, want %d", len(entries), len(want))
	}
	for i := range want {
		if entries[i].Version != want[i] {
			t.Fatalf("entries[%d].Version = %q, want %q", i, entries[i].Version, want[i])
		}
	}
}

func TestRunAllDetailedAndLedgerQuery(t *testing.T) {
	ledger := NewLedger(filepath.Join(t.TempDir(), "migrations.jsonl"))
	runner := NewRunner(ledger)

	migs := []Migration{
		{Version: "20260401_001", Description: "one", Run: func() error { return nil }},
		{Version: "20260401_002", Description: "two", Run: func() error { return nil }},
	}

	applied, diag, err := runner.RunAllDetailed(migs)
	if err != nil {
		t.Fatalf("RunAllDetailed() error = %v", err)
	}
	if len(applied) != 2 || diag.AppliedCount != 2 || diag.DurationMs < 0 {
		t.Fatalf("unexpected diagnostics: applied=%v diag=%+v", applied, diag)
	}

	last, err := ledger.LastVersion()
	if err != nil {
		t.Fatalf("LastVersion() error = %v", err)
	}
	if last != "20260401_002" {
		t.Fatalf("LastVersion() = %q, want 20260401_002", last)
	}

	entries, err := ledger.Query("20260401_", 1)
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	if len(entries) != 1 || entries[0].Version != "20260401_002" {
		t.Fatalf("unexpected query result: %+v", entries)
	}
}

func TestRunAllDetailedFailureDiagnostics(t *testing.T) {
	ledger := NewLedger(filepath.Join(t.TempDir(), "migrations.jsonl"))
	runner := NewRunner(ledger)

	_, diag, err := runner.RunAllDetailed([]Migration{
		{Version: "20260402_001", Run: func() error { return nil }},
		{Version: "20260402_002", Run: func() error { return fmt.Errorf("failed") }},
	})
	if err == nil {
		t.Fatalf("expected error")
	}
	if diag.FailedVersion != "20260402_002" || diag.AppliedCount != 1 {
		t.Fatalf("unexpected diagnostics: %+v", diag)
	}
}

func TestRunnerRunAllDetailedRejectsInvalidMigration(t *testing.T) {
	ledger := NewLedger(filepath.Join(t.TempDir(), "migrations.jsonl"))
	runner := NewRunner(ledger)

	_, _, err := runner.RunAllDetailed([]Migration{{Version: "", Run: func() error { return nil }}})
	if err == nil {
		t.Fatalf("expected version validation error")
	}

	_, _, err = runner.RunAllDetailed([]Migration{{Version: "20260403_001", Run: nil}})
	if err == nil {
		t.Fatalf("expected nil run validation error")
	}
}
