package providers

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestPolicyRespectsMinimumContextClass(t *testing.T) {
	p := testEngine()
	route := types.ModelRoute{Provider: "acme", Primary: "fast-lite", Fast: "fast-lite"}

	decision := p.Select(route, RoutingRequest{Hints: RoutingHints{MinimumContextClass: ContextWindowClassMedium}})
	if decision.Model == "fast-lite" {
		t.Fatalf("expected context class-aware fallback, got %s/%s", decision.Provider, decision.Model)
	}
}
