package providers

import "testing"

func TestMetadataQueryExtendedLookups(t *testing.T) {
	maxOut, ok := LookupModelMaxOutput("openai", "gpt-4o")
	if !ok || maxOut <= 0 {
		t.Fatalf("expected max output lookup, got (%d, %v)", maxOut, ok)
	}
	class, ok := LookupModelCostClass("ollama", "llama3")
	if !ok || class != CostClassFree {
		t.Fatalf("expected free cost class, got (%q, %v)", class, ok)
	}
	tier, ok := LookupModelServiceTier("anthropic", "claude-opus-4-20250514")
	if !ok || tier == "" {
		t.Fatalf("expected service tier lookup, got (%q, %v)", tier, ok)
	}
}
