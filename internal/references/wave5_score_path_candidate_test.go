package references

import "testing"

func TestWave5ScorePathCandidatePrefersExactOverFuzzy(t *testing.T) {
	exact, _, okExact := scorePathCandidate("internal/tui/app.go", "internal/tui/app.go")
	fuzzy, _, okFuzzy := scorePathCandidate("internal/tui/app.go", "itap")
	if !okExact || !okFuzzy {
		t.Fatalf("expected both exact and fuzzy to match")
	}
	if exact <= fuzzy {
		t.Fatalf("expected exact score greater than fuzzy score, got exact=%d fuzzy=%d", exact, fuzzy)
	}
}
