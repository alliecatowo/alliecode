package state

import "sort"

type StartupHydrationState struct {
	SessionID         string `json:"session_id,omitempty"`
	ProjectPath       string `json:"project_path,omitempty"`
	HydrationMode     string `json:"hydration_mode,omitempty"`
	Hydrated          bool   `json:"hydrated"`
	HydratedAtUnix    int64  `json:"hydrated_at_unix,omitempty"`
	CacheFresh        bool   `json:"cache_fresh"`
	CacheScope        string `json:"cache_scope,omitempty"`
	CacheLayeringMode string `json:"cache_layering_mode,omitempty"`
	CachePersistedAt  int64  `json:"cache_persisted_at_unix_nano,omitempty"`
	CacheInvalidates  int    `json:"cache_invalidates,omitempty"`
	ValidationPassed  bool   `json:"validation_passed"`
	ValidationErrors  int    `json:"validation_errors,omitempty"`
	ValidationWarns   int    `json:"validation_warns,omitempty"`
	ValidationSummary string `json:"validation_summary,omitempty"`
	MigrationVersion  string `json:"migration_version,omitempty"`
	MigrationApplied  int    `json:"migration_applied,omitempty"`
	MigrationSkipped  int    `json:"migration_skipped,omitempty"`
	RuntimeSurface    string `json:"runtime_surface,omitempty"`
	CommandSurface    string `json:"command_surface,omitempty"`
	LastError         string `json:"last_error,omitempty"`
	StartupCount      int    `json:"startup_count,omitempty"`
}

type StartupHydrationQuery struct {
	ProjectContains  string
	HydrationMode    string
	Hydrated         *bool
	Validation       *bool
	RequireError     bool
	RequireCache     bool
	RequireMigration bool
	MinStartupCount  int
	Limit            int
}

func BuildStartupHydrationStates(runtimeStates []RuntimeState, cacheItems []SettingsCacheMetadata, diagItems []ConfigDiagnosticsState, query StartupHydrationQuery) []StartupHydrationState {
	if len(runtimeStates) == 0 {
		return nil
	}

	projectContains := normalizeContains(query.ProjectContains)
	hydrationMode := normalizeEquals(query.HydrationMode)

	out := make([]StartupHydrationState, 0, len(runtimeStates))
	for _, runtimeState := range runtimeStates {
		item := StartupHydrationState{
			SessionID:      runtimeState.SessionID,
			ProjectPath:    runtimeState.ProjectPath,
			HydrationMode:  runtimeHydrationMode(runtimeState),
			Hydrated:       runtimeState.HydratedFromDisk || runtimeState.HydratedFromEnvironment,
			HydratedAtUnix: runtimeState.HydratedAtUnix,
			RuntimeSurface: runtimeState.RuntimeSurface,
			CommandSurface: runtimeState.CommandSurface,
			LastError:      runtimeState.LastError,
			StartupCount:   runtimeState.StartupCount,
		}

		cache, hasCache := firstCacheForProject(cacheItems, runtimeState.ProjectPath)
		if hasCache {
			item.CacheFresh = cache.Fresh
			item.CacheScope = cache.LastScope
			item.CacheLayeringMode = cache.LastKnownLayeringMode
			item.CachePersistedAt = cache.LastPersistedAtUnixNano
			item.CacheInvalidates = cache.InvalidateCount
		}

		diag, hasDiag := firstDiagnosticsForProject(diagItems, runtimeState.ProjectPath)
		if hasDiag {
			item.ValidationPassed = diag.LastValidationPassed
			item.ValidationErrors = diag.LastValidationErrorCount
			item.ValidationWarns = diag.LastValidationWarningCount
			item.ValidationSummary = diag.LastDiagnosticsSummary
			item.MigrationVersion = diag.LastMigrationVersion
			item.MigrationApplied = diag.MigrationAppliedCount
			item.MigrationSkipped = diag.MigrationSkippedCount
		}

		if !matchesContains(item.ProjectPath, projectContains) {
			continue
		}
		if !matchesEquals(item.HydrationMode, hydrationMode) {
			continue
		}
		if query.Hydrated != nil && item.Hydrated != *query.Hydrated {
			continue
		}
		if query.Validation != nil && item.ValidationPassed != *query.Validation {
			continue
		}
		if query.RequireError && normalizeContains(item.LastError) == "" {
			continue
		}
		if query.RequireCache && normalizeContains(item.CacheScope) == "" {
			continue
		}
		if query.RequireMigration && normalizeContains(item.MigrationVersion) == "" {
			continue
		}
		if query.MinStartupCount > 0 && item.StartupCount < query.MinStartupCount {
			continue
		}
		out = append(out, item)
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].HydratedAtUnix != out[j].HydratedAtUnix {
			return out[i].HydratedAtUnix > out[j].HydratedAtUnix
		}
		if out[i].StartupCount != out[j].StartupCount {
			return out[i].StartupCount > out[j].StartupCount
		}
		return out[i].SessionID < out[j].SessionID
	})

	limit := clampLimit(query.Limit, len(out))
	return out[:limit]
}

func firstCacheForProject(items []SettingsCacheMetadata, projectPath string) (SettingsCacheMetadata, bool) {
	for _, item := range items {
		if cleanPath(item.LastProjectDir) == cleanPath(projectPath) {
			return item, true
		}
	}
	return SettingsCacheMetadata{}, false
}

func firstDiagnosticsForProject(items []ConfigDiagnosticsState, projectPath string) (ConfigDiagnosticsState, bool) {
	for _, item := range items {
		if cleanPath(item.LastProjectPath) == cleanPath(projectPath) {
			return item, true
		}
	}
	return ConfigDiagnosticsState{}, false
}
