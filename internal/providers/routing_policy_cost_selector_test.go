package providers

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestPolicyCostSelectorExact(t *testing.T) {
	p := testEngine()
	route := types.ModelRoute{Provider: "acme", Primary: "balanced-pro", Fast: "fast-lite"}

	decision := p.Select(route, RoutingRequest{Hints: RoutingHints{AllowedCostClass: CostClassLow, CostClassSelector: "exact"}})
	if decision.Model != "fast-lite" {
		t.Fatalf("expected exact low-cost model, got %s/%s", decision.Provider, decision.Model)
	}
}
