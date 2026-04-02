package config

import "testing"

func TestQueryDiagnosticsFiltersAndLimits(t *testing.T) {
	items := []Diagnostic{
		{Severity: DiagnosticInfo, Code: "a_info", Message: "all good", Remediation: "noop"},
		{Severity: DiagnosticWarn, Code: "b_warn", Message: "provider missing", Remediation: "set provider"},
		{Severity: DiagnosticError, Code: "c_error", Message: "model missing", Remediation: "set model"},
	}

	out := QueryDiagnostics(items, DiagnosticQuery{
		Severity:        DiagnosticWarn,
		CodePrefix:      "b_",
		MessageContains: "missing",
		RequireRemedy:   true,
		IncludeRemedy:   true,
		Limit:           1,
		CaseInsensitive: true,
	})
	if len(out) != 1 {
		t.Fatalf("len(out) = %d, want 1", len(out))
	}
	if out[0].Code != "b_warn" {
		t.Fatalf("unexpected code: %+v", out[0])
	}
}

func TestQueryDiagnosticsCanStripRemediation(t *testing.T) {
	items := []Diagnostic{{Severity: DiagnosticInfo, Code: "x", Message: "m", Remediation: "r"}}
	out := QueryDiagnostics(items, DiagnosticQuery{IncludeRemedy: false})
	if len(out) != 1 || out[0].Remediation != "" {
		t.Fatalf("unexpected output: %+v", out)
	}
}
