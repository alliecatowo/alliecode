package providers

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestPolicyCanDisableGlobalProviderFallback(t *testing.T) {
	p := testEngine()
	route := types.ModelRoute{Provider: "acme", Primary: "balanced-pro", Fast: "fast-lite"}
	decision := p.Select(route, RoutingRequest{Hints: RoutingHints{RequiresAudio: true, DisableProviderFallback: true}})
	if decision.Provider != "acme" {
		t.Fatalf("expected decision to stay on acme, got %s", decision.Provider)
	}
}
