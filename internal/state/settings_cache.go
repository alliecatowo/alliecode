package state

import "strings"

type SettingsCacheMetadata struct {
	CacheFilePath            string `json:"cache_file_path,omitempty"`
	LastProjectDir           string `json:"last_project_dir,omitempty"`
	LastScope                string `json:"last_scope,omitempty"`
	LastLoadUnixNano         int64  `json:"last_load_unix_nano,omitempty"`
	LastHitUnixNano          int64  `json:"last_hit_unix_nano,omitempty"`
	CacheHitCount            int    `json:"cache_hit_count,omitempty"`
	CacheMissCount           int    `json:"cache_miss_count,omitempty"`
	LastInvalidateCause      string `json:"last_invalidate_cause,omitempty"`
	LastInvalidateAt         int64  `json:"last_invalidate_at_unix_nano,omitempty"`
	Fresh                    bool   `json:"fresh,omitempty"`
	LastKnownGlobalPath      string `json:"last_known_global_path,omitempty"`
	LastKnownProjectPath     string `json:"last_known_project_path,omitempty"`
	LastKnownEnvFingerprint  string `json:"last_known_env_fingerprint,omitempty"`
	LastKnownLayeringMode    string `json:"last_known_layering_mode,omitempty"`
	InvalidateCount          int    `json:"invalidate_count,omitempty"`
	ConsecutiveFreshHitCount int    `json:"consecutive_fresh_hit_count,omitempty"`
	ConsecutiveMissCount     int    `json:"consecutive_miss_count,omitempty"`
	LastPersistedAtUnixNano  int64  `json:"last_persisted_at_unix_nano,omitempty"`
	LastError                string `json:"last_error,omitempty"`
	LastErrorAtUnixNano      int64  `json:"last_error_at_unix_nano,omitempty"`
}

func ReadSettingsCacheMetadata(path string) (SettingsCacheMetadata, error) {
	var state SettingsCacheMetadata
	if err := readJSONState(path, "settings cache metadata", &state); err != nil {
		return SettingsCacheMetadata{}, err
	}
	return normalizeSettingsCacheMetadata(state), nil
}

func WriteSettingsCacheMetadata(path string, state SettingsCacheMetadata) error {
	state = normalizeSettingsCacheMetadata(state)
	return writeJSONState(path, "settings cache metadata", state)
}

func normalizeSettingsCacheMetadata(state SettingsCacheMetadata) SettingsCacheMetadata {
	state.CacheFilePath = strings.TrimSpace(state.CacheFilePath)
	state.LastProjectDir = strings.TrimSpace(state.LastProjectDir)
	state.LastScope = strings.TrimSpace(state.LastScope)
	if state.LastLoadUnixNano < 0 {
		state.LastLoadUnixNano = 0
	}
	if state.LastHitUnixNano < 0 {
		state.LastHitUnixNano = 0
	}
	if state.CacheHitCount < 0 {
		state.CacheHitCount = 0
	}
	if state.CacheMissCount < 0 {
		state.CacheMissCount = 0
	}
	state.LastInvalidateCause = strings.TrimSpace(state.LastInvalidateCause)
	if state.LastInvalidateAt < 0 {
		state.LastInvalidateAt = 0
	}
	state.LastKnownGlobalPath = strings.TrimSpace(state.LastKnownGlobalPath)
	state.LastKnownProjectPath = strings.TrimSpace(state.LastKnownProjectPath)
	state.LastKnownEnvFingerprint = strings.TrimSpace(state.LastKnownEnvFingerprint)
	state.LastKnownLayeringMode = strings.TrimSpace(state.LastKnownLayeringMode)
	if state.InvalidateCount < 0 {
		state.InvalidateCount = 0
	}
	if state.ConsecutiveFreshHitCount < 0 {
		state.ConsecutiveFreshHitCount = 0
	}
	if state.ConsecutiveMissCount < 0 {
		state.ConsecutiveMissCount = 0
	}
	if state.LastPersistedAtUnixNano < 0 {
		state.LastPersistedAtUnixNano = 0
	}
	state.LastError = strings.TrimSpace(state.LastError)
	if state.LastErrorAtUnixNano < 0 {
		state.LastErrorAtUnixNano = 0
	}
	return state
}
