package providers

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestPolicyFallbackPriorityRouteBeforeGlobal(t *testing.T) {
	p := testEngine()
	route := types.ModelRoute{Provider: "acme", Primary: "balanced-pro", Fast: "fast-lite"}

	decision := p.Select(route, RoutingRequest{Hints: RoutingHints{
		RequiresAudio:      true,
		FallbackPriorities: []string{"route", "global"},
	}})
	if decision.Reason != "route fallback" {
		t.Fatalf("expected route fallback priority, got reason %q", decision.Reason)
	}
}
