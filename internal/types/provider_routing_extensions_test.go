package types

import "testing"

func TestProviderHintsExtendedRoutingFields(t *testing.T) {
	h := ProviderHints{CostClass: "low", CostClassSelector: "exact", ContextClass: "large", FallbackPriority: "global"}
	if h.CostClassSelector != "exact" || h.ContextClass != "large" {
		t.Fatalf("unexpected provider hints: %+v", h)
	}
}

func TestRoutingPolicyExtendedSelectors(t *testing.T) {
	r := RoutingPolicy{
		FallbackPriorities:  []string{"provider", "global", "route"},
		MinimumContextClass: "large",
		AllowedCostClasses:  []string{"low", "medium"},
		CostClassSelector:   "exact",
	}
	if len(r.FallbackPriorities) != 3 || r.MinimumContextClass != "large" {
		t.Fatalf("unexpected routing policy: %+v", r)
	}
}
