package providers

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestPolicyRespectsServiceTierHint(t *testing.T) {
	p := testEngine()
	route := types.ModelRoute{Provider: "acme"}
	decision := p.Select(route, RoutingRequest{Hints: RoutingHints{ServiceTier: ServiceTierStandard}})
	if decision.Model != "fast-lite" {
		t.Fatalf("expected standard tier fast-lite, got %s", decision.Model)
	}
}
