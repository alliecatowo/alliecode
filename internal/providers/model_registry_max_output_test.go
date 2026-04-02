package providers

import "testing"

func TestModelRegistryMaxOutput(t *testing.T) {
	r := NewDefaultModelMetadataRegistry()
	out, ok := r.MaxOutput("openai", "o3")
	if !ok || out <= 0 {
		t.Fatalf("expected max output for openai/o3, got (%d, %v)", out, ok)
	}
}
