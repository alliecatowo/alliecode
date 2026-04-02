package migrations

import (
	"testing"
	"time"
)

func TestFilterEntries(t *testing.T) {
	entries := []LedgerEntry{
		{Version: "20260401_001", Description: "seed defaults", Status: "applied", AppliedAt: time.Unix(1, 0).UTC()},
		{Version: "20260401_002", Description: "add query", Status: "applied", AppliedAt: time.Unix(2, 0).UTC()},
	}
	out := FilterEntries(entries, Query{VersionPrefix: "20260401_00", Description: "query", Status: "applied", Limit: 5})
	if len(out) != 1 || out[0].Version != "20260401_002" {
		t.Fatalf("unexpected filter result: %+v", out)
	}
}
