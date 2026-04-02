package state

import "sort"

type ConfigDiagnosticsRepository struct {
	path string
}

type ConfigDiagnosticsQuery struct {
	LayeringMode       string
	ValidationPassed   *bool
	MinErrorCount      int
	MinWarningCount    int
	SummaryContains    string
	MigrationVersion   string
	ValidatedAfterUnix int64
	Limit              int
}

type ConfigMigrationReport struct {
	ConfigPath        string `json:"config_path,omitempty"`
	ProjectPath       string `json:"project_path,omitempty"`
	LayeringMode      string `json:"layering_mode,omitempty"`
	MigrationVersion  string `json:"migration_version,omitempty"`
	AppliedCount      int    `json:"applied_count,omitempty"`
	SkippedCount      int    `json:"skipped_count,omitempty"`
	ValidationErrors  int    `json:"validation_errors,omitempty"`
	ValidationWarns   int    `json:"validation_warns,omitempty"`
	ValidationPassed  bool   `json:"validation_passed"`
	ValidationSummary string `json:"validation_summary,omitempty"`
	ValidatedAtUnix   int64  `json:"validated_at_unix,omitempty"`
}

type ConfigMigrationQuery struct {
	MigrationVersion string
	ProjectContains  string
	ConfigContains   string
	LayeringMode     string
	MinApplied       int
	MinSkipped       int
	MinErrors        int
	MinWarnings      int
	ValidationPassed *bool
	Limit            int
}

func NewConfigDiagnosticsRepository(path string) *ConfigDiagnosticsRepository {
	return &ConfigDiagnosticsRepository{path: path}
}

func (r *ConfigDiagnosticsRepository) Path() string {
	if r == nil {
		return ""
	}
	return r.path
}

func (r *ConfigDiagnosticsRepository) Get() (ConfigDiagnosticsState, error) {
	return ReadConfigDiagnosticsState(r.path)
}

func (r *ConfigDiagnosticsRepository) Save(st ConfigDiagnosticsState) error {
	return WriteConfigDiagnosticsState(r.path, st)
}

func (r *ConfigDiagnosticsRepository) Matches(query ConfigDiagnosticsQuery) (bool, ConfigDiagnosticsState, error) {
	st, err := r.Get()
	if err != nil {
		return false, ConfigDiagnosticsState{}, err
	}
	return matchesConfigDiagnosticsQuery(st, query), st, nil
}

func QueryConfigDiagnostics(states []ConfigDiagnosticsState, query ConfigDiagnosticsQuery) []ConfigDiagnosticsState {
	if len(states) == 0 {
		return nil
	}

	layeringMode := normalizeContains(query.LayeringMode)
	summaryContains := normalizeContains(query.SummaryContains)
	migrationVersion := normalizeContains(query.MigrationVersion)

	out := make([]ConfigDiagnosticsState, 0, len(states))
	for _, st := range states {
		if layeringMode != "" && !matchesContains(st.LayeringMode, layeringMode) {
			continue
		}
		if query.ValidationPassed != nil && st.LastValidationPassed != *query.ValidationPassed {
			continue
		}
		if query.MinErrorCount > 0 && st.LastValidationErrorCount < query.MinErrorCount {
			continue
		}
		if query.MinWarningCount > 0 && st.LastValidationWarningCount < query.MinWarningCount {
			continue
		}
		if !matchesContains(st.LastDiagnosticsSummary, summaryContains) {
			continue
		}
		if migrationVersion != "" && !matchesContains(st.LastMigrationVersion, migrationVersion) {
			continue
		}
		if query.ValidatedAfterUnix > 0 && st.LastValidatedAtUnix < query.ValidatedAfterUnix {
			continue
		}
		out = append(out, st)
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].LastValidatedAtUnix != out[j].LastValidatedAtUnix {
			return out[i].LastValidatedAtUnix > out[j].LastValidatedAtUnix
		}
		return out[i].LastConfigPath < out[j].LastConfigPath
	})

	limit := clampLimit(query.Limit, len(out))
	return out[:limit]
}

func matchesConfigDiagnosticsQuery(st ConfigDiagnosticsState, query ConfigDiagnosticsQuery) bool {
	matched := QueryConfigDiagnostics([]ConfigDiagnosticsState{st}, query)
	return len(matched) == 1
}

func BuildConfigMigrationReports(states []ConfigDiagnosticsState, query ConfigMigrationQuery) []ConfigMigrationReport {
	if len(states) == 0 {
		return nil
	}
	migrationVersion := normalizeContains(query.MigrationVersion)
	projectContains := normalizeContains(query.ProjectContains)
	configContains := normalizeContains(query.ConfigContains)
	layeringMode := normalizeContains(query.LayeringMode)

	out := make([]ConfigMigrationReport, 0, len(states))
	for _, st := range states {
		if migrationVersion != "" && !matchesContains(st.LastMigrationVersion, migrationVersion) {
			continue
		}
		if layeringMode != "" && !matchesContains(st.LayeringMode, layeringMode) {
			continue
		}
		if !matchesContains(st.LastProjectPath, projectContains) {
			continue
		}
		if !matchesContains(st.LastConfigPath, configContains) {
			continue
		}
		if query.MinApplied > 0 && st.MigrationAppliedCount < query.MinApplied {
			continue
		}
		if query.MinSkipped > 0 && st.MigrationSkippedCount < query.MinSkipped {
			continue
		}
		if query.MinErrors > 0 && st.LastValidationErrorCount < query.MinErrors {
			continue
		}
		if query.MinWarnings > 0 && st.LastValidationWarningCount < query.MinWarnings {
			continue
		}
		if query.ValidationPassed != nil && st.LastValidationPassed != *query.ValidationPassed {
			continue
		}
		out = append(out, ConfigMigrationReport{
			ConfigPath:        st.LastConfigPath,
			ProjectPath:       st.LastProjectPath,
			LayeringMode:      st.LayeringMode,
			MigrationVersion:  st.LastMigrationVersion,
			AppliedCount:      st.MigrationAppliedCount,
			SkippedCount:      st.MigrationSkippedCount,
			ValidationErrors:  st.LastValidationErrorCount,
			ValidationWarns:   st.LastValidationWarningCount,
			ValidationPassed:  st.LastValidationPassed,
			ValidationSummary: st.LastDiagnosticsSummary,
			ValidatedAtUnix:   st.LastValidatedAtUnix,
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].ValidatedAtUnix != out[j].ValidatedAtUnix {
			return out[i].ValidatedAtUnix > out[j].ValidatedAtUnix
		}
		if out[i].AppliedCount != out[j].AppliedCount {
			return out[i].AppliedCount > out[j].AppliedCount
		}
		return out[i].ConfigPath < out[j].ConfigPath
	})

	limit := clampLimit(query.Limit, len(out))
	return out[:limit]
}
