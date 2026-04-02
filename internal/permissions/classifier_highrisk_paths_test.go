package permissions

import "testing"

func TestClassifyCommandDetailedHighRisk(t *testing.T) {
	classification := ClassifyCommandDetailed("rm -rf /")
	if classification.Level < RiskCritical {
		t.Fatalf("expected critical classification, got %v", classification.Level)
	}
}
