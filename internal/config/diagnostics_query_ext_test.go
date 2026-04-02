package config

import "testing"

func TestQueryDiagnosticsSeverityThresholdAndRemedyContains(t *testing.T) {
	items := []Diagnostic{
		{Severity: DiagnosticInfo, Code: "i", Message: "ok", Remediation: "none"},
		{Severity: DiagnosticWarn, Code: "w", Message: "warn", Remediation: "set provider"},
		{Severity: DiagnosticError, Code: "e", Message: "err", Remediation: "set model"},
	}
	out := QueryDiagnostics(items, DiagnosticQuery{SeverityAtLeast: DiagnosticWarn, RemedyContains: "provider", CaseInsensitive: true, IncludeRemedy: true})
	if len(out) != 1 || out[0].Code != "w" {
		t.Fatalf("unexpected diagnostics query output: %+v", out)
	}
}

func TestQueryDiagnosticsPageSurfaceOffsets(t *testing.T) {
	items := []Diagnostic{{Severity: DiagnosticError, Code: "a", Message: "alpha"}, {Severity: DiagnosticWarn, Code: "b", Message: "bravo"}}
	page := QueryDiagnosticsPage(items, DiagnosticPageQuery{DiagnosticQuery: DiagnosticQuery{}, CodePrefixAny: []string{"b"}, Offset: 0, Limit: 1})
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].Code != "b" {
		t.Fatalf("unexpected paged diagnostics: %+v", page)
	}
}
