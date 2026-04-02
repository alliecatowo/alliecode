package state

import "strings"

type RuntimeState struct {
	SessionID               string `json:"session_id,omitempty"`
	SessionPath             string `json:"session_path,omitempty"`
	ProjectPath             string `json:"project_path,omitempty"`
	WorkingDirectory        string `json:"working_directory,omitempty"`
	RemoteControlAtStartup  bool   `json:"remote_control_at_startup,omitempty"`
	HydratedFromDisk        bool   `json:"hydrated_from_disk,omitempty"`
	HydratedFromEnvironment bool   `json:"hydrated_from_environment,omitempty"`
	HydratedAtUnix          int64  `json:"hydrated_at_unix,omitempty"`
	RuntimeSurface          string `json:"runtime_surface,omitempty"`
	CommandSurface          string `json:"command_surface,omitempty"`
	StartupCount            int    `json:"startup_count,omitempty"`
	LastError               string `json:"last_error,omitempty"`
	LastErrorAtUnix         int64  `json:"last_error_at_unix,omitempty"`
}

func ReadRuntimeState(path string) (RuntimeState, error) {
	var st RuntimeState
	if err := readJSONState(path, "runtime state", &st); err != nil {
		return RuntimeState{}, err
	}
	return normalizeRuntimeState(st), nil
}

func WriteRuntimeState(path string, st RuntimeState) error {
	st = normalizeRuntimeState(st)
	return writeJSONState(path, "runtime state", st)
}

func normalizeRuntimeState(st RuntimeState) RuntimeState {
	st.SessionID = strings.TrimSpace(st.SessionID)
	st.SessionPath = strings.TrimSpace(st.SessionPath)
	st.ProjectPath = strings.TrimSpace(st.ProjectPath)
	st.WorkingDirectory = strings.TrimSpace(st.WorkingDirectory)
	st.RuntimeSurface = strings.TrimSpace(st.RuntimeSurface)
	st.CommandSurface = strings.TrimSpace(st.CommandSurface)
	st.LastError = strings.TrimSpace(st.LastError)
	if st.HydratedAtUnix < 0 {
		st.HydratedAtUnix = 0
	}
	if st.StartupCount < 0 {
		st.StartupCount = 0
	}
	if st.LastErrorAtUnix < 0 {
		st.LastErrorAtUnix = 0
	}
	return st
}
