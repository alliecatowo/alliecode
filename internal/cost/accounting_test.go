package cost

import (
	"math"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestEstimateCostUSD(t *testing.T) {
	usage := types.Usage{InputTokens: 1_200_000, OutputTokens: 300_000}
	model := types.Model{CostPerMInput: 2.0, CostPerMOutput: 8.0}

	got := EstimateCostUSD(usage, model)
	want := (1.2 * 2.0) + (0.3 * 8.0)
	if !almostEqual(got, want) {
		t.Fatalf("EstimateCostUSD() = %f, want %f", got, want)
	}
}

func TestAccountantAggregatesPerSessionAndSummary(t *testing.T) {
	a := NewAccountant()
	model := &types.Model{CostPerMInput: 1.0, CostPerMOutput: 3.0}

	a.AddUsage(UsageRecord{
		SessionID:    "s1",
		Usage:        types.Usage{InputTokens: 100_000, OutputTokens: 200_000, CacheHits: 1},
		TimestampUTC: time.Now().UTC().Add(-2 * time.Minute),
	}, model)
	a.AddUsage(UsageRecord{
		SessionID:    "s1",
		Usage:        types.Usage{InputTokens: 50_000, OutputTokens: 50_000},
		TimestampUTC: time.Now().UTC().Add(-1 * time.Minute),
	}, model)
	a.AddUsage(UsageRecord{
		SessionID:    "s2",
		Usage:        types.Usage{InputTokens: 100_000, OutputTokens: 0},
		TimestampUTC: time.Now().UTC(),
	}, model)

	s1 := a.SessionTotal("s1")
	if s1.Calls != 2 {
		t.Fatalf("s1.Calls = %d, want 2", s1.Calls)
	}
	if s1.InputTokens != 150_000 || s1.OutputTokens != 250_000 {
		t.Fatalf("s1 token totals mismatch: %+v", s1)
	}
	if s1.CacheHits != 1 {
		t.Fatalf("s1.CacheHits = %d, want 1", s1.CacheHits)
	}

	summary := a.Summary()
	if summary.Sessions != 2 {
		t.Fatalf("summary.Sessions = %d, want 2", summary.Sessions)
	}
	if summary.Calls != 3 {
		t.Fatalf("summary.Calls = %d, want 3", summary.Calls)
	}
	if summary.InputTokens != 250_000 || summary.OutputTokens != 250_000 {
		t.Fatalf("summary token totals mismatch: %+v", summary)
	}

	recent := a.RecentSessionTotals(1)
	if len(recent) != 1 {
		t.Fatalf("len(recent) = %d, want 1", len(recent))
	}
	if recent[0].SessionID != "s2" {
		t.Fatalf("recent[0].SessionID = %q, want %q", recent[0].SessionID, "s2")
	}
}

func TestAccountantUsesDirectCostWhenProvided(t *testing.T) {
	a := NewAccountant()
	a.AddUsage(UsageRecord{SessionID: "s3", Usage: types.Usage{InputTokens: 10}, CostUSD: 0.42}, nil)

	got := a.SessionTotal("s3")
	if !almostEqual(got.EstimatedUSD, 0.42) {
		t.Fatalf("EstimatedUSD = %f, want 0.42", got.EstimatedUSD)
	}
}

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}
