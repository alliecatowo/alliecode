package config

import "testing"

func TestQueryLayerDiagnosticsPageRemoteMode(t *testing.T) {
	requireErr := false
	page := QueryLayerDiagnosticsPage([]LayerDiagnostics{
		{EffectivePath: "/a", RemoteMode: "local", ValidationError: ""},
		{EffectivePath: "/b", RemoteMode: "disabled", ValidationError: "bad"},
	}, LayerPageQuery{RemoteMode: "local", RequireValidationErr: &requireErr})
	if page.Total != 1 || page.Items[0].EffectivePath != "/a" {
		t.Fatalf("unexpected layer diagnostics page: %+v", page)
	}
}
