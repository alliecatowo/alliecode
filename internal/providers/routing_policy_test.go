package providers

import (
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func testRegistry() *StaticModelMetadataRegistry {
	return NewStaticModelMetadataRegistry(
		ModelMetadata{Provider: "acme", Model: "fast-lite", Tier: ComplexityTierFast, CostClass: CostClassLow, ServiceTier: ServiceTierStandard, ContextWindow: 8192, SupportsText: true, SupportsTools: true, SupportsToolUse: true},
		ModelMetadata{Provider: "acme", Model: "balanced-pro", Tier: ComplexityTierBalanced, CostClass: CostClassMedium, ServiceTier: ServiceTierPremium, ContextWindow: 128000, SupportsText: true, SupportsTools: true, SupportsToolUse: true},
		ModelMetadata{Provider: "acme", Model: "reasoner-x", Tier: ComplexityTierReasoning, CostClass: CostClassHigh, ServiceTier: ServiceTierPremium, ContextWindow: 200000, SupportsText: true, SupportsTools: true, SupportsToolUse: true, SupportsThinking: true},
		ModelMetadata{Provider: "acme", Model: "visual-v", Tier: ComplexityTierBalanced, CostClass: CostClassMedium, ServiceTier: ServiceTierPremium, ContextWindow: 128000, SupportsText: true, SupportsImage: true, SupportsTools: true, SupportsToolUse: true, SupportsVision: true},
		ModelMetadata{Provider: "beta", Model: "audio-max", Tier: ComplexityTierBalanced, CostClass: CostClassMedium, ServiceTier: ServiceTierPremium, ContextWindow: 256000, SupportsText: true, SupportsAudio: true, SupportsAttachments: true, SupportsTools: true, SupportsToolUse: true},
	)
}

func testEngine() *PolicyEngine {
	r := testRegistry()
	return NewPolicyEngine(r, NewRegistryCapabilityMatrix(r))
}

func TestCapabilityMatrixSupportsRequirements(t *testing.T) {
	r := testRegistry()
	m := NewRegistryCapabilityMatrix(r)

	if !m.Supports("acme", "fast-lite", CapabilityRequirements{NeedsTools: true}) {
		t.Fatalf("expected tools support")
	}
	if m.Supports("acme", "fast-lite", CapabilityRequirements{NeedsThinking: true}) {
		t.Fatalf("expected thinking capability to be gated")
	}
	if m.Supports("acme", "fast-lite", CapabilityRequirements{NeedsVision: true}) {
		t.Fatalf("expected vision capability to be gated")
	}
	if m.Supports("acme", "fast-lite", CapabilityRequirements{MinContextWindow: 16000}) {
		t.Fatalf("expected context window requirement to be gated")
	}
	if !m.Supports("acme", "visual-v", CapabilityRequirements{NeedsImage: true, NeedsVision: true}) {
		t.Fatalf("expected image and vision support")
	}
	if m.Supports("acme", "fast-lite", CapabilityRequirements{NeedsAudio: true}) {
		t.Fatalf("expected audio capability to be gated")
	}
	if !m.Supports("beta", "audio-max", CapabilityRequirements{NeedsAudio: true, NeedsAttachments: true, NeedsToolUse: true}) {
		t.Fatalf("expected multimodal capability requirements to pass")
	}
}

func TestPolicySelectsFastModelForSimpleTask(t *testing.T) {
	p := testEngine()
	route := types.ModelRoute{Provider: "acme", Primary: "balanced-pro", Fast: "fast-lite"}

	decision := p.Select(route, RoutingRequest{Complexity: TaskComplexitySimple})
	if decision.Provider != "acme" || decision.Model != "fast-lite" {
		t.Fatalf("expected acme/fast-lite, got %s/%s", decision.Provider, decision.Model)
	}
}

func TestPolicyPrefersReasoningForComplexTask(t *testing.T) {
	p := testEngine()
	route := types.ModelRoute{Provider: "acme", Primary: "balanced-pro", Fast: "fast-lite"}

	decision := p.Select(route, RoutingRequest{Complexity: TaskComplexityComplex})
	if decision.Provider != "acme" || decision.Model != "reasoner-x" {
		t.Fatalf("expected acme/reasoner-x, got %s/%s", decision.Provider, decision.Model)
	}
}

func TestPolicyRespectsExplicitModelHintWhenCapable(t *testing.T) {
	p := testEngine()
	route := types.ModelRoute{Primary: "balanced-pro"}

	decision := p.Select(route, RoutingRequest{Hints: RoutingHints{Model: "reasoner-x", RequiresThinking: true}})
	if decision.Provider != "acme" || decision.Model != "reasoner-x" {
		t.Fatalf("expected acme/reasoner-x, got %s/%s", decision.Provider, decision.Model)
	}
}

func TestPolicyFallsBackWhenHintModelCannotMeetCapability(t *testing.T) {
	p := testEngine()
	route := types.ModelRoute{Provider: "acme", Primary: "balanced-pro", Fast: "fast-lite"}

	decision := p.Select(route, RoutingRequest{Hints: RoutingHints{Model: "fast-lite", RequiresThinking: true}})
	if decision.Provider != "acme" || decision.Model != "reasoner-x" {
		t.Fatalf("expected capability fallback to acme/reasoner-x, got %s/%s", decision.Provider, decision.Model)
	}
}

func TestPolicyFallsBackAcrossProvidersForCapability(t *testing.T) {
	p := testEngine()
	route := types.ModelRoute{Provider: "acme", Primary: "balanced-pro", Fast: "fast-lite"}

	decision := p.Select(route, RoutingRequest{Hints: RoutingHints{RequiresAudio: true, RequiresAttachments: true}})
	if decision.Provider != "beta" || decision.Model != "audio-max" {
		t.Fatalf("expected global capability fallback to beta/audio-max, got %s/%s", decision.Provider, decision.Model)
	}
}
