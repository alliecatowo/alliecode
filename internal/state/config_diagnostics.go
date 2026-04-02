package state

import "strings"

type ConfigDiagnosticsState struct {
	LastConfigPath             string `json:"last_config_path,omitempty"`
	LastProjectPath            string `json:"last_project_path,omitempty"`
	LayeringMode               string `json:"layering_mode,omitempty"`
	LastValidationPassed       bool   `json:"last_validation_passed,omitempty"`
	LastValidationErrorCount   int    `json:"last_validation_error_count,omitempty"`
	LastValidationWarningCount int    `json:"last_validation_warning_count,omitempty"`
	LastDiagnosticsSummary     string `json:"last_diagnostics_summary,omitempty"`
	LastValidatedAtUnix        int64  `json:"last_validated_at_unix,omitempty"`
	LastMigrationVersion       string `json:"last_migration_version,omitempty"`
	MigrationAppliedCount      int    `json:"migration_applied_count,omitempty"`
	MigrationSkippedCount      int    `json:"migration_skipped_count,omitempty"`
}

func ReadConfigDiagnosticsState(path string) (ConfigDiagnosticsState, error) {
	var st ConfigDiagnosticsState
	if err := readJSONState(path, "config diagnostics", &st); err != nil {
		return ConfigDiagnosticsState{}, err
	}
	return normalizeConfigDiagnosticsState(st), nil
}

func WriteConfigDiagnosticsState(path string, st ConfigDiagnosticsState) error {
	st = normalizeConfigDiagnosticsState(st)
	return writeJSONState(path, "config diagnostics", st)
}

func normalizeConfigDiagnosticsState(st ConfigDiagnosticsState) ConfigDiagnosticsState {
	st.LastConfigPath = strings.TrimSpace(st.LastConfigPath)
	st.LastProjectPath = strings.TrimSpace(st.LastProjectPath)
	st.LayeringMode = strings.TrimSpace(st.LayeringMode)
	st.LastDiagnosticsSummary = strings.TrimSpace(st.LastDiagnosticsSummary)
	st.LastMigrationVersion = strings.TrimSpace(st.LastMigrationVersion)
	if st.LastValidationErrorCount < 0 {
		st.LastValidationErrorCount = 0
	}
	if st.LastValidationWarningCount < 0 {
		st.LastValidationWarningCount = 0
	}
	if st.LastValidatedAtUnix < 0 {
		st.LastValidatedAtUnix = 0
	}
	if st.MigrationAppliedCount < 0 {
		st.MigrationAppliedCount = 0
	}
	if st.MigrationSkippedCount < 0 {
		st.MigrationSkippedCount = 0
	}
	return st
}
