package state

import "strings"

type StartupHydrationCompatibility struct {
	SessionID         string `json:"session_id,omitempty"`
	ProjectPath       string `json:"project_path,omitempty"`
	ConfiguredMode    string `json:"configured_mode,omitempty"`
	RuntimeMode       string `json:"runtime_mode,omitempty"`
	EffectiveMode     string `json:"effective_mode,omitempty"`
	RuntimeSurface    string `json:"runtime_surface,omitempty"`
	CommandSurface    string `json:"command_surface,omitempty"`
	RequiresCompat    bool   `json:"requires_compat"`
	CompatibilityHint string `json:"compatibility_hint,omitempty"`
}

type StartupHydrationCompatibilityQuery struct {
	ProjectContains string
	SessionContains string
	ConfiguredMode  string
	RuntimeMode     string
	EffectiveMode   string
	RequireCompat   *bool
	Offset          int
	Limit           int
}

func NormalizeHydrationMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "default", "auto":
		return "compat"
	case "legacy", "disk":
		return "disk"
	case "env", "environment":
		return "env"
	case "disk+env", "env+disk", "hybrid":
		return "disk+env"
	case "strict", "cold":
		return "cold"
	case "compat":
		return "compat"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func ResolveHydrationCompatibilityMode(configured string, runtime RuntimeState) string {
	runtimeMode := NormalizeHydrationMode(runtimeHydrationMode(runtime))
	configuredMode := NormalizeHydrationMode(configured)
	if configuredMode == "compat" {
		return runtimeMode
	}
	if configuredMode == "cold" {
		return "cold"
	}
	if configuredMode == "disk" && runtime.HydratedFromEnvironment && !runtime.HydratedFromDisk {
		return "env"
	}
	if configuredMode == "env" && runtime.HydratedFromDisk && !runtime.HydratedFromEnvironment {
		return "disk"
	}
	return runtimeMode
}

func BuildStartupHydrationCompatibility(states []RuntimeState, configuredMode string, query StartupHydrationCompatibilityQuery) []StartupHydrationCompatibility {
	if len(states) == 0 {
		return nil
	}
	projectContains := normalizeContains(query.ProjectContains)
	sessionContains := normalizeContains(query.SessionContains)
	configuredModeFilter := normalizeEquals(query.ConfiguredMode)
	runtimeMode := normalizeEquals(query.RuntimeMode)
	effectiveMode := normalizeEquals(query.EffectiveMode)

	out := make([]StartupHydrationCompatibility, 0, len(states))
	for _, st := range states {
		resolved := ResolveHydrationCompatibilityMode(configuredMode, st)
		item := StartupHydrationCompatibility{
			SessionID:      st.SessionID,
			ProjectPath:    st.ProjectPath,
			ConfiguredMode: NormalizeHydrationMode(configuredMode),
			RuntimeMode:    NormalizeHydrationMode(runtimeHydrationMode(st)),
			EffectiveMode:  resolved,
			RuntimeSurface: st.RuntimeSurface,
			CommandSurface: st.CommandSurface,
		}
		item.RequiresCompat = item.ConfiguredMode != "compat" && item.ConfiguredMode != item.RuntimeMode
		if item.RequiresCompat {
			item.CompatibilityHint = "runtime hydration differs from configured startup mode"
		}

		if !matchesContains(item.ProjectPath, projectContains) {
			continue
		}
		if !matchesContains(item.SessionID, sessionContains) {
			continue
		}
		if configuredModeFilter != "" && !matchesEquals(item.ConfiguredMode, configuredModeFilter) {
			continue
		}
		if runtimeMode != "" && !matchesEquals(item.RuntimeMode, runtimeMode) {
			continue
		}
		if effectiveMode != "" && !matchesEquals(item.EffectiveMode, effectiveMode) {
			continue
		}
		if query.RequireCompat != nil && item.RequiresCompat != *query.RequireCompat {
			continue
		}
		out = append(out, item)
	}

	start, end := paginateBounds(len(out), query.Offset, query.Limit)
	return out[start:end]
}
