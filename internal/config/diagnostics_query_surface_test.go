package config

import "testing"

func TestQueryDiagnosticsPageAndSummary(t *testing.T) {
	items := []Diagnostic{
		{Severity: DiagnosticError, Code: "config_invalid", Message: "bad provider", Remediation: "fix"},
		{Severity: DiagnosticWarn, Code: "provider_default_missing", Message: "provider empty"},
		{Severity: DiagnosticInfo, Code: "model_default_missing", Message: "model empty", Remediation: "set model"},
	}
	requireRemedy := true
	page := QueryDiagnosticsPage(items, DiagnosticPageQuery{DiagnosticQuery: DiagnosticQuery{IncludeRemedy: true}, Codes: []string{"config_invalid", "provider_default_missing", "model_default_missing"}, CodePrefixAny: []string{"config", "model"}, RemedyContainsAny: []string{"set"}, RequireRemedy: &requireRemedy, Offset: 0, Limit: 1})
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].Code != "model_default_missing" {
		t.Fatalf("unexpected page: %+v", page)
	}
	sum := SummarizeDiagnostics(items)
	if sum.Total != 3 || sum.Errors != 1 || sum.CodePrefixHits["config"] != 1 || sum.DistinctCodes != 3 {
		t.Fatalf("unexpected summary: %+v", sum)
	}
}
