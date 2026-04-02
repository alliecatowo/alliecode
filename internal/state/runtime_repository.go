package state

import "sort"

type RuntimeStateRepository struct {
	path string
}

type RuntimeQuery struct {
	SessionIDContains   string
	ProjectPathContains string
	RuntimeSurface      string
	CommandSurface      string
	HydratedFromDisk    *bool
	HydratedFromEnv     *bool
	MinimumStartupCount int
	MaximumStartupCount int
	LastErrorContains   string
	HydratedAfterUnix   int64
	Limit               int
}

type RuntimeHydrationSummary struct {
	SessionID               string `json:"session_id,omitempty"`
	ProjectPath             string `json:"project_path,omitempty"`
	HydrationMode           string `json:"hydration_mode,omitempty"`
	Hydrated                bool   `json:"hydrated"`
	HydratedFromDisk        bool   `json:"hydrated_from_disk"`
	HydratedFromEnvironment bool   `json:"hydrated_from_environment"`
	HydratedAtUnix          int64  `json:"hydrated_at_unix,omitempty"`
	StartupCount            int    `json:"startup_count,omitempty"`
	LastError               string `json:"last_error,omitempty"`
	RuntimeSurface          string `json:"runtime_surface,omitempty"`
	CommandSurface          string `json:"command_surface,omitempty"`
}

type RuntimeHydrationQuery struct {
	HydrationMode   string
	ProjectContains string
	SessionContains string
	Hydrated        *bool
	WithErrorsOnly  bool
	MinStartupCount int
	Limit           int
}

func NewRuntimeStateRepository(path string) *RuntimeStateRepository {
	return &RuntimeStateRepository{path: path}
}

func (r *RuntimeStateRepository) Get() (RuntimeState, error) {
	return ReadRuntimeState(r.path)
}

func (r *RuntimeStateRepository) Save(st RuntimeState) error {
	return WriteRuntimeState(r.path, st)
}

func QueryRuntimeStates(states []RuntimeState, query RuntimeQuery) []RuntimeState {
	if len(states) == 0 {
		return nil
	}
	sessionContains := normalizeContains(query.SessionIDContains)
	projectContains := normalizeContains(query.ProjectPathContains)
	runtimeSurface := normalizeContains(query.RuntimeSurface)
	commandSurface := normalizeContains(query.CommandSurface)
	lastErrorContains := normalizeContains(query.LastErrorContains)

	out := make([]RuntimeState, 0, len(states))
	for _, st := range states {
		if !matchesContains(st.SessionID, sessionContains) {
			continue
		}
		if !matchesContains(st.ProjectPath, projectContains) {
			continue
		}
		if runtimeSurface != "" && normalizeContains(st.RuntimeSurface) != runtimeSurface {
			continue
		}
		if commandSurface != "" && normalizeContains(st.CommandSurface) != commandSurface {
			continue
		}
		if query.HydratedFromDisk != nil && st.HydratedFromDisk != *query.HydratedFromDisk {
			continue
		}
		if query.HydratedFromEnv != nil && st.HydratedFromEnvironment != *query.HydratedFromEnv {
			continue
		}
		if query.MinimumStartupCount > 0 && st.StartupCount < query.MinimumStartupCount {
			continue
		}
		if query.MaximumStartupCount > 0 && st.StartupCount > query.MaximumStartupCount {
			continue
		}
		if !matchesContains(st.LastError, lastErrorContains) {
			continue
		}
		if query.HydratedAfterUnix > 0 && st.HydratedAtUnix < query.HydratedAfterUnix {
			continue
		}
		out = append(out, st)
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].HydratedAtUnix != out[j].HydratedAtUnix {
			return out[i].HydratedAtUnix > out[j].HydratedAtUnix
		}
		return out[i].SessionID < out[j].SessionID
	})

	limit := clampLimit(query.Limit, len(out))
	return out[:limit]
}

func BuildRuntimeHydrationSummaries(states []RuntimeState, query RuntimeHydrationQuery) []RuntimeHydrationSummary {
	if len(states) == 0 {
		return nil
	}
	hydrationMode := normalizeEquals(query.HydrationMode)
	projectContains := normalizeContains(query.ProjectContains)
	sessionContains := normalizeContains(query.SessionContains)

	out := make([]RuntimeHydrationSummary, 0, len(states))
	for _, st := range states {
		summary := RuntimeHydrationSummary{
			SessionID:               st.SessionID,
			ProjectPath:             st.ProjectPath,
			HydratedFromDisk:        st.HydratedFromDisk,
			HydratedFromEnvironment: st.HydratedFromEnvironment,
			HydratedAtUnix:          st.HydratedAtUnix,
			StartupCount:            st.StartupCount,
			LastError:               st.LastError,
			RuntimeSurface:          st.RuntimeSurface,
			CommandSurface:          st.CommandSurface,
		}
		summary.Hydrated = st.HydratedFromDisk || st.HydratedFromEnvironment
		summary.HydrationMode = runtimeHydrationMode(st)

		if !matchesContains(summary.ProjectPath, projectContains) {
			continue
		}
		if !matchesContains(summary.SessionID, sessionContains) {
			continue
		}
		if !matchesEquals(summary.HydrationMode, hydrationMode) {
			continue
		}
		if query.Hydrated != nil && summary.Hydrated != *query.Hydrated {
			continue
		}
		if query.WithErrorsOnly && normalizeContains(summary.LastError) == "" {
			continue
		}
		if query.MinStartupCount > 0 && summary.StartupCount < query.MinStartupCount {
			continue
		}
		out = append(out, summary)
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

func BuildRuntimeHydrationCompatibility(states []RuntimeState, configuredMode string, query StartupHydrationCompatibilityQuery) []StartupHydrationCompatibility {
	return BuildStartupHydrationCompatibility(states, configuredMode, query)
}

func runtimeHydrationMode(st RuntimeState) string {
	if st.HydratedFromDisk && st.HydratedFromEnvironment {
		return "disk+env"
	}
	if st.HydratedFromDisk {
		return "disk"
	}
	if st.HydratedFromEnvironment {
		return "env"
	}
	return "cold"
}
