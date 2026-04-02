package types

import "testing"

func TestModelMetadataExtendedFields(t *testing.T) {
	m := ModelMetadata{Provider: "openai", Model: "o3", Tier: "reasoning", ServiceTier: "premium", CostClass: "high"}
	if m.ServiceTier != "premium" || m.CostClass != "high" {
		t.Fatalf("unexpected metadata fields: %+v", m)
	}
}
