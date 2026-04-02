package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigDiagnostics_ReportsValidationError(t *testing.T) {
	cfg := &Config{DefaultProvider: "", DefaultModel: ""}
	diags := ConfigDiagnostics(cfg)
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics for invalid config")
	}
	if !hasDiagnostic(diags, "config_invalid", DiagnosticError) {
		t.Fatalf("expected config_invalid error diagnostic")
	}
	if !hasDiagnosticWithRemediation(diags, "config_invalid", DiagnosticError) {
		t.Fatalf("expected config_invalid diagnostic to include remediation")
	}
	if !hasDiagnostic(diags, "config_invalid_detail", DiagnosticError) {
		t.Fatalf("expected detailed validation diagnostics")
	}
}

func TestConfigDiagnostics_CompanionMutedInheritedInfo(t *testing.T) {
	cfg := NewDefaultConfig()
	if cfg.CompanionMuted != nil {
		t.Fatalf("expected default companion muted to be nil")
	}
	diags := ConfigDiagnostics(cfg)
	if !hasDiagnostic(diags, "companion_muted_inherited", DiagnosticInfo) {
		t.Fatalf("expected companion muted inheritance diagnostic")
	}
}

func TestPathDiagnostics_FileStates(t *testing.T) {
	tmp := t.TempDir()
	globalPath := filepath.Join(tmp, "global.yaml")
	projectPath := filepath.Join(tmp, "project.conf")

	if err := os.WriteFile(globalPath, []byte("default_provider: openai\n"), 0o644); err != nil {
		t.Fatalf("write global file: %v", err)
	}
	if err := os.WriteFile(projectPath, []byte("default_provider: openai\n"), 0o644); err != nil {
		t.Fatalf("write project file: %v", err)
	}

	diags := PathDiagnostics(globalPath, projectPath)
	if !hasDiagnostic(diags, "global_config_ok", DiagnosticInfo) {
		t.Fatalf("expected global ok diagnostic")
	}
	if !hasDiagnostic(diags, "project_config_extension_unusual", DiagnosticWarn) {
		t.Fatalf("expected extension warning diagnostic")
	}
	if !hasDiagnosticWithRemediation(diags, "project_config_extension_unusual", DiagnosticWarn) {
		t.Fatalf("expected extension warning remediation")
	}
}

func TestFormatDiagnostics_StableOrderingAndRendering(t *testing.T) {
	diags := []Diagnostic{
		{Severity: DiagnosticInfo, Code: "z_info", Message: "info msg", Remediation: "later"},
		{Severity: DiagnosticWarn, Code: "a_warn", Message: "warn msg", Remediation: "soon"},
		{Severity: DiagnosticError, Code: "b_error", Message: "error msg", Remediation: "now"},
		{Severity: DiagnosticError, Code: "a_error", Message: "first error", Remediation: "urgent"},
	}

	rendered := FormatDiagnostics(diags)
	want := strings.Join([]string{
		"[error] a_error: first error (remediation: urgent)",
		"[error] b_error: error msg (remediation: now)",
		"[warn] a_warn: warn msg (remediation: soon)",
		"[info] z_info: info msg (remediation: later)",
	}, "\n")

	if rendered != want {
		t.Fatalf("formatted diagnostics mismatch\nwant:\n%s\n\ngot:\n%s", want, rendered)
	}
}

func TestFormatDiagnostics_EmptyInput(t *testing.T) {
	if got := FormatDiagnostics(nil); got != "" {
		t.Fatalf("expected empty output for nil diagnostics, got %q", got)
	}
}

func TestLayerDiagnosticsToDiagnosticsAndCounts(t *testing.T) {
	items := LayerDiagnosticsToDiagnostics(LayerDiagnostics{
		LayeringMode:     "layered_project",
		DefaultProvider:  "openai",
		DefaultModel:     "gpt-4o-mini",
		ValidationPassed: false,
		ValidationError:  "boom",
	})
	if len(items) == 0 {
		t.Fatalf("expected layer diagnostics")
	}
	errs, warns := ValidationCounts(items)
	if errs == 0 {
		t.Fatalf("expected at least one error, got errs=%d warns=%d", errs, warns)
	}
}

func TestQueryDiagnosticsSortedBySeverityAndCode(t *testing.T) {
	items := []Diagnostic{
		{Severity: DiagnosticInfo, Code: "z", Message: "i"},
		{Severity: DiagnosticWarn, Code: "b", Message: "w"},
		{Severity: DiagnosticError, Code: "a", Message: "e"},
	}
	out := QueryDiagnostics(items, DiagnosticQuery{SortBySeverity: true, SortByCode: true, IncludeRemedy: true})
	if len(out) != 3 {
		t.Fatalf("len(out) = %d", len(out))
	}
	if out[0].Code != "a" || out[1].Code != "b" || out[2].Code != "z" {
		t.Fatalf("unexpected order: %+v", out)
	}
}

func hasDiagnostic(diags []Diagnostic, code string, sev DiagnosticSeverity) bool {
	for _, diag := range diags {
		if diag.Code == code && diag.Severity == sev {
			return true
		}
	}
	return false
}

func hasDiagnosticWithRemediation(diags []Diagnostic, code string, sev DiagnosticSeverity) bool {
	for _, diag := range diags {
		if diag.Code == code && diag.Severity == sev && strings.TrimSpace(diag.Remediation) != "" {
			return true
		}
	}
	return false
}
