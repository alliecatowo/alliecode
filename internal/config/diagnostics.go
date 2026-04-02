package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type DiagnosticSeverity string

const (
	DiagnosticInfo  DiagnosticSeverity = "info"
	DiagnosticWarn  DiagnosticSeverity = "warn"
	DiagnosticError DiagnosticSeverity = "error"
)

type Diagnostic struct {
	Severity    DiagnosticSeverity `json:"severity"`
	Code        string             `json:"code"`
	Message     string             `json:"message"`
	Remediation string             `json:"remediation,omitempty"`
}

func ConfigDiagnostics(cfg *Config) []Diagnostic {
	if cfg == nil {
		return []Diagnostic{{Severity: DiagnosticError, Code: "config_nil", Message: "configuration is nil", Remediation: "load a configuration before running diagnostics"}}
	}

	diags := make([]Diagnostic, 0, 8)
	if err := cfg.Validate(); err != nil {
		diags = append(diags, Diagnostic{Severity: DiagnosticError, Code: "config_invalid", Message: err.Error(), Remediation: "fix invalid fields in config and rerun diagnostics"})
		for _, item := range flattenValidationErrors(err) {
			diags = append(diags, Diagnostic{Severity: DiagnosticError, Code: "config_invalid_detail", Message: item, Remediation: "update the referenced field in the highest-precedence config source"})
		}
	}

	if strings.TrimSpace(cfg.DefaultProvider) == "" {
		diags = append(diags, Diagnostic{Severity: DiagnosticWarn, Code: "provider_default_missing", Message: "default provider is not set", Remediation: "set default_provider in config or ALLIECODE_DEFAULT_PROVIDER"})
	}
	if strings.TrimSpace(cfg.DefaultModel) == "" {
		diags = append(diags, Diagnostic{Severity: DiagnosticWarn, Code: "model_default_missing", Message: "default model is not set", Remediation: "set default_model in config or ALLIECODE_DEFAULT_MODEL"})
	}
	if len(cfg.Providers) == 0 {
		diags = append(diags, Diagnostic{Severity: DiagnosticInfo, Code: "providers_empty", Message: "no provider credentials configured", Remediation: "add provider credentials under providers.<name>"})
	}
	if cfg.CompanionMuted == nil {
		diags = append(diags, Diagnostic{Severity: DiagnosticInfo, Code: "companion_muted_inherited", Message: "companion_muted is unset and inherits runtime default", Remediation: "set companion_muted in config or ALLIECODE_COMPANION_MUTED to force a value"})
	}
	if strings.TrimSpace(cfg.Startup.HydrationMode) == "" {
		diags = append(diags, Diagnostic{Severity: DiagnosticInfo, Code: "startup_hydration_mode_defaulted", Message: "startup.hydration_mode is unset and defaults to compat", Remediation: "set startup.hydration_mode to legacy, compat, or strict"})
	}
	if cfg.MigrationVersion == 0 {
		diags = append(diags, Diagnostic{Severity: DiagnosticInfo, Code: "migration_version_defaulted", Message: "migration_version is unset and defaults to startup baseline", Remediation: "set migration_version in config to suppress baseline migration checks"})
	}

	return diags
}

func LayerDiagnosticsToDiagnostics(info LayerDiagnostics) []Diagnostic {
	items := make([]Diagnostic, 0, 10)
	if strings.TrimSpace(info.EffectivePath) == "" {
		items = append(items, Diagnostic{Severity: DiagnosticWarn, Code: "effective_path_missing", Message: "effective config path is empty", Remediation: "load config with explicit path or LoadLayered"})
	} else {
		items = append(items, Diagnostic{Severity: DiagnosticInfo, Code: "effective_path", Message: "effective config path is " + info.EffectivePath, Remediation: "no action needed"})
	}

	items = append(items,
		Diagnostic{Severity: DiagnosticInfo, Code: "layering_mode", Message: "layering mode is " + safeOr(info.LayeringMode, "unknown"), Remediation: "no action needed"},
		Diagnostic{Severity: DiagnosticInfo, Code: "default_provider", Message: "default provider is " + safeOr(info.DefaultProvider, "<unset>"), Remediation: "set default_provider if unset"},
		Diagnostic{Severity: DiagnosticInfo, Code: "default_model", Message: "default model is " + safeOr(info.DefaultModel, "<unset>"), Remediation: "set default_model if unset"},
	)

	if info.ProviderFromEnv {
		items = append(items, Diagnostic{Severity: DiagnosticInfo, Code: "provider_from_env", Message: "default provider comes from environment", Remediation: "unset ALLIECODE_DEFAULT_PROVIDER to use file config"})
	}
	if info.ModelFromEnv {
		items = append(items, Diagnostic{Severity: DiagnosticInfo, Code: "model_from_env", Message: "default model comes from environment", Remediation: "unset ALLIECODE_DEFAULT_MODEL to use file config"})
	}
	if !info.ValidationPassed {
		items = append(items, Diagnostic{Severity: DiagnosticError, Code: "effective_config_invalid", Message: safeOr(info.ValidationError, "validation failed"), Remediation: "fix invalid fields in highest-precedence source"})
	}
	return items
}

func ValidationCounts(diags []Diagnostic) (errorsCount, warnsCount int) {
	for _, diag := range diags {
		switch diag.Severity {
		case DiagnosticError:
			errorsCount++
		case DiagnosticWarn:
			warnsCount++
		}
	}
	return errorsCount, warnsCount
}

func flattenValidationErrors(err error) []string {
	if err == nil {
		return nil
	}
	collector := make([]string, 0, 4)
	var walk func(error)
	walk = func(current error) {
		if current == nil {
			return
		}
		type unwrapMany interface{ Unwrap() []error }
		if multi, ok := current.(unwrapMany); ok {
			for _, nested := range multi.Unwrap() {
				walk(nested)
			}
			return
		}
		collector = append(collector, current.Error())
	}
	walk(err)
	if len(collector) == 0 {
		return []string{err.Error()}
	}
	return collector
}

func PathDiagnostics(globalPath, projectPath string) []Diagnostic {
	diags := make([]Diagnostic, 0, 6)
	diags = append(diags, fileDiagnostic(globalPath, "global")...)
	diags = append(diags, fileDiagnostic(projectPath, "project")...)
	return diags
}

func fileDiagnostic(path, kind string) []Diagnostic {
	if strings.TrimSpace(path) == "" {
		return []Diagnostic{{Severity: DiagnosticWarn, Code: kind + "_path_missing", Message: fmt.Sprintf("%s config path is empty", kind), Remediation: fmt.Sprintf("set %s config path before loading layered config", kind)}}
	}

	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Diagnostic{{Severity: DiagnosticInfo, Code: kind + "_config_missing", Message: fmt.Sprintf("%s config not found at %s", kind, path), Remediation: fmt.Sprintf("create %s config at %s", kind, path)}}
		}
		return []Diagnostic{{Severity: DiagnosticError, Code: kind + "_config_unreadable", Message: fmt.Sprintf("cannot stat %s config %s: %v", kind, path, err), Remediation: "check file permissions and parent directory access"}}
	}

	if st.IsDir() {
		return []Diagnostic{{Severity: DiagnosticError, Code: kind + "_config_is_directory", Message: fmt.Sprintf("%s config path %s is a directory", kind, path), Remediation: "replace directory with a YAML file path"}}
	}

	if filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".yml" {
		return []Diagnostic{{Severity: DiagnosticWarn, Code: kind + "_config_extension_unusual", Message: fmt.Sprintf("%s config path %s has non-YAML extension", kind, path), Remediation: "rename file to .yaml or .yml for clarity"}}
	}

	return []Diagnostic{{Severity: DiagnosticInfo, Code: kind + "_config_ok", Message: fmt.Sprintf("%s config found at %s", kind, path), Remediation: "no action needed"}}
}

func FormatDiagnostics(diags []Diagnostic) string {
	if len(diags) == 0 {
		return ""
	}

	ordered := make([]Diagnostic, len(diags))
	copy(ordered, diags)
	sort.SliceStable(ordered, func(i, j int) bool {
		left := severityRank(ordered[i].Severity)
		right := severityRank(ordered[j].Severity)
		if left != right {
			return left > right
		}
		if ordered[i].Code != ordered[j].Code {
			return ordered[i].Code < ordered[j].Code
		}
		if ordered[i].Message != ordered[j].Message {
			return ordered[i].Message < ordered[j].Message
		}
		return ordered[i].Remediation < ordered[j].Remediation
	})

	lines := make([]string, 0, len(ordered))
	for _, diag := range ordered {
		line := fmt.Sprintf("[%s] %s: %s", diag.Severity, diag.Code, diag.Message)
		if strings.TrimSpace(diag.Remediation) != "" {
			line += " (remediation: " + diag.Remediation + ")"
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func severityRank(severity DiagnosticSeverity) int {
	switch severity {
	case DiagnosticError:
		return 3
	case DiagnosticWarn:
		return 2
	case DiagnosticInfo:
		return 1
	default:
		return 0
	}
}

func safeOr(v, fallback string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return fallback
	}
	return v
}

func LayeringSummary(info LayerDiagnostics) string {
	parts := []string{
		"mode=" + safeOr(info.LayeringMode, "unknown"),
		"provider=" + safeOr(info.DefaultProvider, "<unset>"),
		"model=" + safeOr(info.DefaultModel, "<unset>"),
		"hydration=" + safeOr(info.HydrationMode, "compat"),
	}
	if info.ValidationPassed {
		parts = append(parts, "validation=passed")
	} else {
		parts = append(parts, "validation=failed")
	}
	return strings.Join(parts, ",")
}
