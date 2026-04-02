package state

import "testing"

func TestQueryConfigDiagnostics(t *testing.T) {
	states := []ConfigDiagnosticsState{
		{LayeringMode: "layered", LastValidationPassed: true, LastValidationErrorCount: 0, LastDiagnosticsSummary: "ok", LastValidatedAtUnix: 1},
		{LayeringMode: "single", LastValidationPassed: false, LastValidationErrorCount: 2, LastDiagnosticsSummary: "bad provider", LastValidatedAtUnix: 2},
	}
	passed := true
	out := QueryConfigDiagnostics(states, ConfigDiagnosticsQuery{LayeringMode: "layer", ValidationPassed: &passed, Limit: 5})
	if len(out) != 1 || !out[0].LastValidationPassed {
		t.Fatalf("unexpected query result: %+v", out)
	}
}
