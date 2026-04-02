package state

import "sort"

type SettingsCacheRepository struct {
	path string
}

type SettingsCacheQuery struct {
	Scope              string
	ProjectContains    string
	LayeringMode       string
	Fresh              *bool
	MinHitCount        int
	MinMissCount       int
	MinInvalidateCount int
	MinFreshHitStreak  int
	InvalidateCause    string
	ErrorContains      string
	PersistedAfter     int64
	Limit              int
}

type SettingsCacheSnapshot struct {
	ProjectDir              string `json:"project_dir,omitempty"`
	Scope                   string `json:"scope,omitempty"`
	LayeringMode            string `json:"layering_mode,omitempty"`
	Fresh                   bool   `json:"fresh"`
	HitCount                int    `json:"hit_count,omitempty"`
	MissCount               int    `json:"miss_count,omitempty"`
	InvalidateCount         int    `json:"invalidate_count,omitempty"`
	ConsecutiveFreshHits    int    `json:"consecutive_fresh_hits,omitempty"`
	ConsecutiveMisses       int    `json:"consecutive_misses,omitempty"`
	LastLoadUnixNano        int64  `json:"last_load_unix_nano,omitempty"`
	LastHitUnixNano         int64  `json:"last_hit_unix_nano,omitempty"`
	EnvFingerprint          string `json:"env_fingerprint,omitempty"`
	LastInvalidateCause     string `json:"last_invalidate_cause,omitempty"`
	LastPersistedAtUnixNano int64  `json:"last_persisted_at_unix_nano,omitempty"`
	LastError               string `json:"last_error,omitempty"`
}

type SettingsCacheSnapshotQuery struct {
	ProjectContains string
	Scope           string
	LayeringMode    string
	Fresh           *bool
	HasError        *bool
	MinInvalidates  int
	Limit           int
}

func NewSettingsCacheRepository(path string) *SettingsCacheRepository {
	return &SettingsCacheRepository{path: path}
}

func (r *SettingsCacheRepository) Get() (SettingsCacheMetadata, error) {
	return ReadSettingsCacheMetadata(r.path)
}

func (r *SettingsCacheRepository) Save(st SettingsCacheMetadata) error {
	return WriteSettingsCacheMetadata(r.path, st)
}

func QuerySettingsCacheMetadata(items []SettingsCacheMetadata, query SettingsCacheQuery) []SettingsCacheMetadata {
	if len(items) == 0 {
		return nil
	}
	scope := normalizeContains(query.Scope)
	projectContains := normalizeContains(query.ProjectContains)
	layeringMode := normalizeContains(query.LayeringMode)
	invalidateCause := normalizeContains(query.InvalidateCause)
	errorContains := normalizeContains(query.ErrorContains)

	out := make([]SettingsCacheMetadata, 0, len(items))
	for _, item := range items {
		if scope != "" && normalizeContains(item.LastScope) != scope {
			continue
		}
		if !matchesContains(item.LastProjectDir, projectContains) {
			continue
		}
		if layeringMode != "" && normalizeContains(item.LastKnownLayeringMode) != layeringMode {
			continue
		}
		if query.Fresh != nil && item.Fresh != *query.Fresh {
			continue
		}
		if query.MinHitCount > 0 && item.CacheHitCount < query.MinHitCount {
			continue
		}
		if query.MinMissCount > 0 && item.CacheMissCount < query.MinMissCount {
			continue
		}
		if query.MinInvalidateCount > 0 && item.InvalidateCount < query.MinInvalidateCount {
			continue
		}
		if query.MinFreshHitStreak > 0 && item.ConsecutiveFreshHitCount < query.MinFreshHitStreak {
			continue
		}
		if invalidateCause != "" && normalizeContains(item.LastInvalidateCause) != invalidateCause {
			continue
		}
		if !matchesContains(item.LastError, errorContains) {
			continue
		}
		if query.PersistedAfter > 0 && item.LastPersistedAtUnixNano < query.PersistedAfter {
			continue
		}
		out = append(out, item)
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].LastPersistedAtUnixNano != out[j].LastPersistedAtUnixNano {
			return out[i].LastPersistedAtUnixNano > out[j].LastPersistedAtUnixNano
		}
		return out[i].LastProjectDir < out[j].LastProjectDir
	})

	limit := clampLimit(query.Limit, len(out))
	return out[:limit]
}

func BuildSettingsCacheSnapshots(items []SettingsCacheMetadata, query SettingsCacheSnapshotQuery) []SettingsCacheSnapshot {
	if len(items) == 0 {
		return nil
	}
	projectContains := normalizeContains(query.ProjectContains)
	scope := normalizeEquals(query.Scope)
	layering := normalizeEquals(query.LayeringMode)

	out := make([]SettingsCacheSnapshot, 0, len(items))
	for _, item := range items {
		hasError := normalizeContains(item.LastError) != ""
		if !matchesContains(item.LastProjectDir, projectContains) {
			continue
		}
		if !matchesEquals(item.LastScope, scope) {
			continue
		}
		if !matchesEquals(item.LastKnownLayeringMode, layering) {
			continue
		}
		if query.Fresh != nil && item.Fresh != *query.Fresh {
			continue
		}
		if query.HasError != nil && hasError != *query.HasError {
			continue
		}
		if query.MinInvalidates > 0 && item.InvalidateCount < query.MinInvalidates {
			continue
		}
		out = append(out, SettingsCacheSnapshot{
			ProjectDir:              item.LastProjectDir,
			Scope:                   item.LastScope,
			LayeringMode:            item.LastKnownLayeringMode,
			Fresh:                   item.Fresh,
			HitCount:                item.CacheHitCount,
			MissCount:               item.CacheMissCount,
			InvalidateCount:         item.InvalidateCount,
			ConsecutiveFreshHits:    item.ConsecutiveFreshHitCount,
			ConsecutiveMisses:       item.ConsecutiveMissCount,
			LastLoadUnixNano:        item.LastLoadUnixNano,
			LastHitUnixNano:         item.LastHitUnixNano,
			EnvFingerprint:          item.LastKnownEnvFingerprint,
			LastInvalidateCause:     item.LastInvalidateCause,
			LastPersistedAtUnixNano: item.LastPersistedAtUnixNano,
			LastError:               item.LastError,
		})
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].LastPersistedAtUnixNano != out[j].LastPersistedAtUnixNano {
			return out[i].LastPersistedAtUnixNano > out[j].LastPersistedAtUnixNano
		}
		if out[i].InvalidateCount != out[j].InvalidateCount {
			return out[i].InvalidateCount > out[j].InvalidateCount
		}
		return out[i].ProjectDir < out[j].ProjectDir
	})

	limit := clampLimit(query.Limit, len(out))
	return out[:limit]
}
