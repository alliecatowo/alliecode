package state

type Repositories struct {
	Paths             Paths
	Projects          *ProjectsRepository
	Sessions          *SessionMetadataRepository
	ConfigDiagnostics *ConfigDiagnosticsRepository
	SettingsCache     *SettingsCacheRepository
	Runtime           *RuntimeStateRepository
	Auth              *AuthStateRepository
	Onboarding        *ProjectOnboardingRepository
}

func NewRepositories(paths Paths) Repositories {
	return Repositories{
		Paths:             paths,
		Projects:          NewProjectsRepository(paths.ProjectsStateFile),
		Sessions:          NewSessionMetadataRepository(paths.SessionMetadataFile),
		ConfigDiagnostics: NewConfigDiagnosticsRepository(paths.ConfigDiagnosticsFile),
		SettingsCache:     NewSettingsCacheRepository(paths.SettingsCacheFile),
		Runtime:           NewRuntimeStateRepository(paths.RuntimeStateFile),
		Auth:              NewAuthStateRepository(paths.AuthStateFile),
		Onboarding:        NewProjectOnboardingRepository(paths.ProjectOnboardingStateFile),
	}
}

func (r Repositories) BuildStartupHydrationSummary(runtimeStates []RuntimeState, cacheItems []SettingsCacheMetadata, diagnostics []ConfigDiagnosticsState, query StartupHydrationQuery) []StartupHydrationState {
	return BuildStartupHydrationStates(runtimeStates, cacheItems, diagnostics, query)
}

func (r Repositories) BuildStartupHydrationCompatibility(runtimeStates []RuntimeState, configuredMode string, query StartupHydrationCompatibilityQuery) []StartupHydrationCompatibility {
	return BuildStartupHydrationCompatibility(runtimeStates, configuredMode, query)
}

func (r Repositories) BuildStartupHydrationSummaryPage(runtimeStates []RuntimeState, cacheItems []SettingsCacheMetadata, diagnostics []ConfigDiagnosticsState, query StartupHydrationQuery, offset, limit int) ([]StartupHydrationState, int) {
	items := BuildStartupHydrationStates(runtimeStates, cacheItems, diagnostics, query)
	total := len(items)
	start, end := paginateBounds(total, offset, limit)
	return items[start:end], total
}
