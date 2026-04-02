package config

import "testing"

func TestFilterLayerDiagnosticsExtendedFields(t *testing.T) {
	fromEnv := true
	items := []LayerDiagnostics{
		{LayeringMode: "layered", DefaultProvider: "openai", RemoteMode: "local", StrictHydration: true, ValidationPassed: true, ProviderFromEnv: true, MigrationVersion: 3},
		{LayeringMode: "single", DefaultProvider: "anthropic", ValidationPassed: true, ProviderFromEnv: false, MigrationVersion: 1},
	}
	out := FilterLayerDiagnostics(items, LayerQuery{ProviderFromEnv: &fromEnv, MinMigration: 2, ValidationState: "passed"})
	if len(out) != 1 || out[0].DefaultProvider != "openai" {
		t.Fatalf("unexpected filtered diagnostics: %+v", out)
	}
}

func TestSummarizeLayerDiagnosticsCounts(t *testing.T) {
	items := []LayerDiagnostics{{LayeringMode: "layered", HydrationMode: "compat", ValidationPassed: true}, {LayeringMode: "single", HydrationMode: "strict", ValidationPassed: false}}
	summary := SummarizeLayerDiagnostics(items)
	if summary.Total != 2 || summary.ValidationFailed != 1 || summary.HydrationModes["compat"] != 1 {
		t.Fatalf("unexpected layer summary: %+v", summary)
	}
}
