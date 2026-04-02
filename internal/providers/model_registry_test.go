package providers

import "testing"

func TestMetadataPricingAndContextQueries(t *testing.T) {
	registry := NewDefaultModelMetadataRegistry()

	ctx, ok := registry.ContextWindow("openai", "gpt-4o")
	if !ok {
		t.Fatalf("expected context window for openai/gpt-4o")
	}
	if ctx != 128000 {
		t.Fatalf("context window = %d, want 128000", ctx)
	}

	in, out, ok := registry.Pricing("anthropic", "claude-sonnet-4-20250514")
	if !ok {
		t.Fatalf("expected pricing for anthropic model")
	}
	if in <= 0 || out <= 0 {
		t.Fatalf("expected positive pricing, got in=%f out=%f", in, out)
	}
}

func TestListByMinimumContextWindow(t *testing.T) {
	registry := NewDefaultModelMetadataRegistry()

	models := registry.ListByMinimumContextWindow("openai", 150000)
	if len(models) == 0 {
		t.Fatalf("expected models with context window >= 150000")
	}
	for _, m := range models {
		if m.ContextWindow > 0 && m.ContextWindow < 150000 {
			t.Fatalf("model %s has context %d, expected >= 150000", m.Model, m.ContextWindow)
		}
	}
}

func TestMetadataQueryHelpers(t *testing.T) {
	meta, ok := LookupModelMetadata("openai", "gpt-4o-mini")
	if !ok {
		t.Fatalf("expected metadata for openai/gpt-4o-mini")
	}
	if meta.ContextWindow != 128000 {
		t.Fatalf("context window = %d, want 128000", meta.ContextWindow)
	}

	ctx, ok := LookupModelContextWindow("openai", "o3")
	if !ok || ctx != 200000 {
		t.Fatalf("context lookup = (%d, %v), want (200000, true)", ctx, ok)
	}

	in, out, ok := LookupModelPricing("openai", "o4-mini")
	if !ok || in <= 0 || out <= 0 {
		t.Fatalf("pricing lookup = (%f, %f, %v), expected positive values", in, out, ok)
	}

	withContext := ListModelsByMinimumContextWindow("anthropic", 180000)
	if len(withContext) == 0 {
		t.Fatalf("expected anthropic models with large context")
	}
}

func TestCapabilitySummaryIncludesMultimodalFlags(t *testing.T) {
	summary, ok := LookupModelCapabilitySummary("openai", "gpt-4o")
	if !ok {
		t.Fatalf("expected capability summary for openai/gpt-4o")
	}
	if summary != "text,image,audio,tool_use,vision,attachments" {
		t.Fatalf("unexpected capability summary: %q", summary)
	}

	ollamaSummary, ok := LookupModelCapabilitySummary("ollama", "llama3")
	if !ok {
		t.Fatalf("expected capability summary for ollama/llama3")
	}
	if ollamaSummary != "text,tool_use" {
		t.Fatalf("unexpected ollama capability summary: %q", ollamaSummary)
	}
}
