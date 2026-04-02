package providers

import "testing"

func TestListByServiceTier(t *testing.T) {
	r := NewDefaultModelMetadataRegistry()
	models := r.ListByServiceTier("anthropic", ServiceTierPremium)
	if len(models) == 0 {
		t.Fatalf("expected premium anthropic models")
	}
	for _, m := range models {
		if m.ServiceTier != ServiceTierPremium {
			t.Fatalf("unexpected tier %q for %s", m.ServiceTier, m.Model)
		}
	}
}
