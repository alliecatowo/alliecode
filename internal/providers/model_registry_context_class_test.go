package providers

import "testing"

func TestModelRegistryContextClassResolution(t *testing.T) {
	r := NewStaticModelMetadataRegistry(
		ModelMetadata{Provider: "x", Model: "tiny", ContextWindow: 4096, SupportsText: true},
		ModelMetadata{Provider: "x", Model: "xlarge", ContextWindow: 1000000, SupportsText: true},
	)

	classTiny, ok := r.ContextWindowClass("x", "tiny")
	if !ok || classTiny != ContextWindowClassTiny {
		t.Fatalf("unexpected tiny class: %q (ok=%v)", classTiny, ok)
	}
	classXL, ok := r.ContextWindowClass("x", "xlarge")
	if !ok || classXL != ContextWindowClassXLarge {
		t.Fatalf("unexpected xlarge class: %q (ok=%v)", classXL, ok)
	}
}

func TestModelRegistryListByMinimumContextClass(t *testing.T) {
	r := NewStaticModelMetadataRegistry(
		ModelMetadata{Provider: "x", Model: "small", ContextWindow: 32768, SupportsText: true},
		ModelMetadata{Provider: "x", Model: "medium", ContextWindow: 200000, SupportsText: true},
	)

	models := r.ListByMinimumContextClass("x", ContextWindowClassMedium)
	if len(models) != 1 || models[0].Model != "medium" {
		t.Fatalf("unexpected context class filter result: %+v", models)
	}
}
