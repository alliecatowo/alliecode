package config

import (
	"strings"
	"testing"
)

func TestLayeringSummary(t *testing.T) {
	summary := LayeringSummary(LayerDiagnostics{LayeringMode: "layered", DefaultProvider: "openai", DefaultModel: "gpt-4o-mini", HydrationMode: "compat", ValidationPassed: true})
	for _, want := range []string{"mode=layered", "provider=openai", "validation=passed"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary %q missing %q", summary, want)
		}
	}
}
