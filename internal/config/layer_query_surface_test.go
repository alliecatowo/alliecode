package config

import "testing"

func TestQueryLayerDiagnosticsPageAndSummary(t *testing.T) {
	items := []LayerDiagnostics{
		{EffectivePath: "/repo/a/.alliecode/config.yaml", LayeringMode: "layered", HydrationMode: "compat", RemoteMode: "local", ValidationPassed: true},
		{EffectivePath: "/repo/b/.alliecode/config.yaml", LayeringMode: "single", HydrationMode: "strict", StrictHydration: true, RemoteMode: "disabled", ValidationPassed: false, ValidationError: "missing provider"},
	}
	requireErr := true
	page := QueryLayerDiagnosticsPage(items, LayerPageQuery{LayerQuery: LayerQuery{}, PathContains: "/repo", ValidationErrContains: "missing", RemoteMode: "disabled", RequireValidationErr: &requireErr, Limit: 5})
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].LayeringMode != "single" {
		t.Fatalf("unexpected page: %+v", page)
	}
	sum := SummarizeLayerDiagnostics(items)
	if sum.Total != 2 || sum.ValidationFailed != 1 || sum.LayeringModes["layered"] != 1 || sum.RemoteModes["local"] != 1 || sum.StrictHydration != 1 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
}
