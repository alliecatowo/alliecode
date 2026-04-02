package providers

import "testing"

func TestMetadataQueryLists(t *testing.T) {
	cheap := ListModelsByCostClass("openai", CostClassLow)
	if len(cheap) == 0 {
		t.Fatalf("expected at least one cheap openai model")
	}
	standard := ListModelsByServiceTier("openai", ServiceTierStandard)
	if len(standard) == 0 {
		t.Fatalf("expected standard tier models")
	}
}
