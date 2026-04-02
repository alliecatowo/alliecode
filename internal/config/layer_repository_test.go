package config

import "testing"

func TestFilterLayerDiagnostics(t *testing.T) {
	items := []LayerDiagnostics{
		{LayeringMode: "layered_project", DefaultProvider: "openai", DefaultModel: "gpt-4o-mini", RemoteMode: "local", HydrationMode: "compat", ValidationPassed: true, MigrationVersion: 2},
		{LayeringMode: "single_file", DefaultProvider: "anthropic", DefaultModel: "claude", RemoteMode: "disabled", HydrationMode: "legacy", ValidationPassed: false, MigrationVersion: 1},
	}
	out := FilterLayerDiagnostics(items, LayerQuery{LayeringMode: "layered", Provider: "openai", ValidationState: "passed", Limit: 5})
	if len(out) != 1 || out[0].DefaultProvider != "openai" {
		t.Fatalf("unexpected filter result: %+v", out)
	}
}
