package types

import "testing"

func TestRoutingPolicyExtendedFields(t *testing.T) {
	r := RoutingPolicy{Strategy: "capability", MinimumContextWindow: 8000, CostClassCeiling: "low", ServiceTier: "standard", RequireCapabilities: []string{"tools"}}
	if r.MinimumContextWindow != 8000 || r.CostClassCeiling != "low" || r.ServiceTier != "standard" {
		t.Fatalf("unexpected routing policy values: %+v", r)
	}
}
