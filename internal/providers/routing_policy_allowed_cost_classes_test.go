package providers

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestPolicyAllowedCostClasses(t *testing.T) {
	p := testEngine()
	route := types.ModelRoute{Provider: "acme", Primary: "reasoner-x", Fast: "fast-lite"}

	decision := p.Select(route, RoutingRequest{Hints: RoutingHints{AllowedCostClasses: []CostClass{CostClassLow, CostClassMedium}}})
	if decision.Model == "reasoner-x" {
		t.Fatalf("expected high-cost model to be filtered out")
	}
}
