package providers

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestPolicyPrefersLowCostCandidate(t *testing.T) {
	p := testEngine()
	route := types.ModelRoute{Provider: "acme"}
	decision := p.Select(route, RoutingRequest{Complexity: TaskComplexityNormal, Hints: RoutingHints{PreferLowCost: true, AllowedCostClass: CostClassLow}})
	if decision.Model != "fast-lite" {
		t.Fatalf("expected low-cost model fast-lite, got %s", decision.Model)
	}
}
