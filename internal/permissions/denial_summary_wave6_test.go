package permissions

import (
	"path/filepath"
	"testing"
)

func TestDenialSummaryEmptyLedgerWave6(t *testing.T) {
	ledger := NewDenialsLedger(filepath.Join(t.TempDir(), "empty.jsonl"))
	summary, err := ledger.Summary(DenialQuery{})
	if err != nil {
		t.Fatalf("summary failed: %v", err)
	}
	if summary.Total != 0 {
		t.Fatalf("expected empty summary: %+v", summary)
	}
}
