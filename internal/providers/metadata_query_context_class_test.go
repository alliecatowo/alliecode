package providers

import "testing"

func TestMetadataQueryContextClassLookups(t *testing.T) {
	class, ok := LookupModelContextClass("gemini", "gemini-2.5-pro")
	if !ok || class == "" {
		t.Fatalf("expected context class lookup, got (%q, %v)", class, ok)
	}

	models := ListModelsByMinimumContextClass("openai", ContextWindowClassMedium)
	if len(models) == 0 {
		t.Fatalf("expected at least one openai model in medium+ class")
	}
}
