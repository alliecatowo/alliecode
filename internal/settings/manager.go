package settings

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alliecatowo/alliecode/internal/config"
	"github.com/alliecatowo/alliecode/internal/state"
)

const defaultFreshnessWindow = 5 * time.Second

type Scope string

const (
	ScopeGlobal  Scope = "global"
	ScopeProject Scope = "project"
)

type LoadOptions struct {
	ProjectDir string
	Scope      *Scope
}

type SaveOptions struct {
	ProjectDir string
	Scope      Scope
}

type InvalidateReason string

const (
	InvalidateReasonManual          InvalidateReason = "manual"
	InvalidateReasonProjectSwitch   InvalidateReason = "project_switch"
	InvalidateReasonSourceChanged   InvalidateReason = "source_changed"
	InvalidateReasonFreshnessWindow InvalidateReason = "freshness_window"
)

type Manager struct {
	mu       sync.RWMutex
	maxAge   time.Duration
	entries  map[string]cacheEntry
	timeline *Timeline
}

type RuntimeHydrationSummary struct {
	ProjectDir           string `json:"project_dir,omitempty"`
	ConfigPath           string `json:"config_path,omitempty"`
	Scope                Scope  `json:"scope,omitempty"`
	DefaultProvider      string `json:"default_provider,omitempty"`
	DefaultModel         string `json:"default_model,omitempty"`
	RemoteMode           string `json:"remote_mode,omitempty"`
	HydrationMode        string `json:"hydration_mode,omitempty"`
	ValidationPassed     bool   `json:"validation_passed"`
	RuntimeSurface       string `json:"runtime_surface,omitempty"`
	CommandSurface       string `json:"command_surface,omitempty"`
	MigrationVersion     int    `json:"migration_version,omitempty"`
	CacheHitCount        int    `json:"cache_hit_count,omitempty"`
	CacheMissCount       int    `json:"cache_miss_count,omitempty"`
	CacheInvalidateCount int    `json:"cache_invalidate_count,omitempty"`
	Fresh                bool   `json:"fresh"`
	LoadedAtUnix         int64  `json:"loaded_at_unix,omitempty"`
}

type RuntimeHydrationQuery struct {
	Scope             *Scope
	Provider          string
	RemoteMode        string
	HydrationMode     string
	ProjectContains   string
	MinMigration      int
	RequireValidation *bool
	Fresh             *bool
	Limit             int
}

type cacheEntry struct {
	cfg            *config.Config
	loadedAt       time.Time
	signatures     []fileSignature
	envFingerprint string
	scopePaths     ScopePaths
}

type fileSignature struct {
	path    string
	exists  bool
	size    int64
	modUnix int64
}

func NewManager(maxAge time.Duration) *Manager {
	if maxAge <= 0 {
		maxAge = defaultFreshnessWindow
	}
	return &Manager{
		maxAge:   maxAge,
		entries:  make(map[string]cacheEntry),
		timeline: NewTimeline(),
	}
}

func (m *Manager) Load(opts LoadOptions) (*config.Config, bool, error) {
	target, err := resolveLoadTarget(opts)
	if err != nil {
		return nil, false, err
	}

	m.mu.RLock()
	entry, found := m.entries[target.key]
	m.mu.RUnlock()
	if found {
		if entry.envFingerprint != buildEnvFingerprint() {
			m.recordInvalidation(target.projectDir, string(InvalidateReasonSourceChanged))
		} else {
			fresh, err := m.entryFresh(target.paths, entry)
			if err != nil {
				return nil, false, err
			}
			if fresh {
				m.recordCacheMetadata(target.projectDir, target.scope, true, true, target.scopePaths, entry.envFingerprint, "")
				return entry.cfg, true, nil
			}
			m.recordInvalidation(target.projectDir, string(InvalidateReasonSourceChanged))
		}
	}

	cfg, err := target.load()
	if err != nil {
		return nil, false, err
	}

	sigs, err := signaturesFor(target.paths)
	if err != nil {
		return nil, false, err
	}

	m.mu.Lock()
	m.entries[target.key] = cacheEntry{
		cfg:            cfg,
		loadedAt:       time.Now(),
		signatures:     sigs,
		envFingerprint: buildEnvFingerprint(),
		scopePaths:     target.scopePaths,
	}
	m.timeline.Add(SnapshotFromConfig(scopeFromString(target.scope), target.projectDir, cfg, time.Now()))
	m.mu.Unlock()
	m.recordCacheMetadata(target.projectDir, target.scope, false, true, target.scopePaths, buildEnvFingerprint(), "")

	return cfg, false, nil
}

func (m *Manager) SnapshotTimeline(query SnapshotQuery) []Snapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.timeline == nil {
		return nil
	}
	return m.timeline.Query(query)
}

func (m *Manager) RuntimeHydrationSummaries(projectDir string, query RuntimeHydrationQuery) ([]RuntimeHydrationSummary, error) {
	timelineQuery := SnapshotQuery{
		Scope:           query.Scope,
		Provider:        query.Provider,
		HydrationMode:   query.HydrationMode,
		ProjectContains: query.ProjectContains,
		MinMigration:    query.MinMigration,
		Limit:           query.Limit,
	}
	items := m.SnapshotTimeline(timelineQuery)
	if len(items) == 0 {
		return nil, nil
	}
	cacheMeta, err := m.ReadCacheMetadata(projectDir)
	if err != nil {
		return nil, err
	}
	out := make([]RuntimeHydrationSummary, 0, len(items))
	for _, item := range items {
		diag := LayerDiagnosticsFromSnapshot(item)
		summary := RuntimeHydrationSummary{
			ProjectDir:           item.ProjectDir,
			ConfigPath:           diag.EffectivePath,
			Scope:                item.Scope,
			DefaultProvider:      item.DefaultProvider,
			DefaultModel:         item.DefaultModel,
			RemoteMode:           diag.RemoteMode,
			HydrationMode:        item.HydrationMode,
			ValidationPassed:     diag.ValidationPassed,
			RuntimeSurface:       item.RuntimeSurface,
			CommandSurface:       item.CommandSurface,
			MigrationVersion:     item.MigrationVersion,
			CacheHitCount:        cacheMeta.CacheHitCount,
			CacheMissCount:       cacheMeta.CacheMissCount,
			CacheInvalidateCount: cacheMeta.InvalidateCount,
			Fresh:                cacheMeta.Fresh,
			LoadedAtUnix:         item.LoadedAt.Unix(),
		}
		if query.Fresh != nil && summary.Fresh != *query.Fresh {
			continue
		}
		if remote := strings.ToLower(strings.TrimSpace(query.RemoteMode)); remote != "" && strings.ToLower(strings.TrimSpace(summary.RemoteMode)) != remote {
			continue
		}
		if query.RequireValidation != nil && summary.ValidationPassed != *query.RequireValidation {
			continue
		}
		out = append(out, summary)
	}
	if query.Limit > 0 && len(out) > query.Limit {
		out = out[len(out)-query.Limit:]
	}
	return out, nil
}

func (m *Manager) BuildStartupHydrationStates(runtimeStates []state.RuntimeState, cacheItems []state.SettingsCacheMetadata, diagItems []state.ConfigDiagnosticsState, query state.StartupHydrationQuery) []state.StartupHydrationState {
	return state.BuildStartupHydrationStates(runtimeStates, cacheItems, diagItems, query)
}

func (m *Manager) RuntimeHydrationCompatibility(states []state.RuntimeState, configuredMode string, query state.StartupHydrationCompatibilityQuery) []state.StartupHydrationCompatibility {
	return state.BuildStartupHydrationCompatibility(states, configuredMode, query)
}

func scopeFromString(scope string) Scope {
	switch strings.TrimSpace(scope) {
	case string(ScopeGlobal):
		return ScopeGlobal
	case string(ScopeProject):
		return ScopeProject
	default:
		return ScopeProject
	}
}

func (m *Manager) CacheFilePath(projectDir string) (string, error) {
	paths, err := state.ResolvePaths("", projectDir)
	if err != nil {
		return "", err
	}
	return paths.SettingsCacheFile, nil
}

func (m *Manager) ReadCacheMetadata(projectDir string) (state.SettingsCacheMetadata, error) {
	path, err := m.CacheFilePath(projectDir)
	if err != nil {
		return state.SettingsCacheMetadata{}, err
	}
	return state.ReadSettingsCacheMetadata(path)
}

func (m *Manager) WriteCacheMetadata(projectDir string, metadata state.SettingsCacheMetadata) error {
	path, err := m.CacheFilePath(projectDir)
	if err != nil {
		return err
	}
	return state.WriteSettingsCacheMetadata(path, metadata)
}

func (m *Manager) IsFresh(opts LoadOptions) (bool, error) {
	target, err := resolveLoadTarget(opts)
	if err != nil {
		return false, err
	}

	m.mu.RLock()
	entry, found := m.entries[target.key]
	m.mu.RUnlock()
	if !found {
		return false, nil
	}

	return m.entryFresh(target.paths, entry)
}

func (m *Manager) Invalidate(opts LoadOptions) error {
	return m.InvalidateWithReason(opts, InvalidateReasonManual)
}

func (m *Manager) InvalidateWithReason(opts LoadOptions, reason InvalidateReason) error {
	target, err := resolveLoadTarget(opts)
	if err != nil {
		return err
	}

	m.mu.Lock()
	delete(m.entries, target.key)
	m.mu.Unlock()
	m.recordInvalidation(target.projectDir, string(reason))
	return nil
}

func (m *Manager) InvalidateAll() {
	m.mu.Lock()
	m.entries = make(map[string]cacheEntry)
	m.mu.Unlock()
}

func (m *Manager) HydrateForRuntime(opts LoadOptions, runtimeSurface, commandSurface string) (*config.Config, bool, error) {
	cfg, hit, err := m.Load(opts)
	if err != nil {
		return nil, false, err
	}
	if cfg == nil {
		return nil, hit, nil
	}
	cloned := config.Merge(cfg)
	cloned.Runtime.Surface = strings.TrimSpace(runtimeSurface)
	cloned.Runtime.CommandSurface = strings.TrimSpace(commandSurface)
	if cloned.Startup.HydrationMode == "" {
		cloned.Startup.HydrationMode = "compat"
	}
	return cloned, hit, nil
}

func (m *Manager) entryFresh(paths []string, entry cacheEntry) (bool, error) {
	if m.maxAge > 0 && time.Since(entry.loadedAt) > m.maxAge {
		return false, nil
	}

	current, err := signaturesFor(paths)
	if err != nil {
		return false, err
	}

	if len(current) != len(entry.signatures) {
		return false, nil
	}

	for i := range current {
		if current[i] != entry.signatures[i] {
			return false, nil
		}
	}

	return true, nil
}

type loadTarget struct {
	key        string
	projectDir string
	scope      string
	paths      []string
	scopePaths ScopePaths
	load       func() (*config.Config, error)
}

func resolveLoadTarget(opts LoadOptions) (loadTarget, error) {
	projectDir := opts.ProjectDir
	if projectDir == "" {
		var err error
		projectDir, err = os.Getwd()
		if err != nil {
			return loadTarget{}, fmt.Errorf("resolve working directory: %w", err)
		}
	}

	scopePaths, err := ResolveScopePaths(opts)
	if err != nil {
		return loadTarget{}, err
	}
	globalPath := scopePaths.GlobalPath
	projectPath := scopePaths.ProjectPath

	if opts.Scope == nil {
		return loadTarget{
			key:        "layered|" + filepath.Clean(projectDir),
			projectDir: projectDir,
			scope:      "layered",
			paths:      []string{globalPath, projectPath},
			scopePaths: scopePaths,
			load: func() (*config.Config, error) {
				return config.LoadLayered(projectDir)
			},
		}, nil
	}

	scopePath, err := ResolvePath(PathOptions{ProjectDir: projectDir, Scope: *opts.Scope})
	if err != nil {
		return loadTarget{}, err
	}

	return loadTarget{
		key:        string(*opts.Scope) + "|" + scopePath,
		projectDir: projectDir,
		scope:      string(*opts.Scope),
		paths:      []string{scopePath},
		scopePaths: scopePaths,
		load: func() (*config.Config, error) {
			return config.Load(scopePath)
		},
	}, nil
}

func (m *Manager) recordCacheMetadata(projectDir, scope string, cacheHit bool, fresh bool, paths ScopePaths, envFingerprint, lastErr string) {
	meta, err := m.ReadCacheMetadata(projectDir)
	if err != nil {
		return
	}
	now := time.Now().UnixNano()
	meta.LastProjectDir = filepath.Clean(projectDir)
	meta.LastScope = scope
	if path, err := m.CacheFilePath(projectDir); err == nil {
		meta.CacheFilePath = path
	}
	meta.LastKnownGlobalPath = paths.GlobalPath
	meta.LastKnownProjectPath = paths.ProjectPath
	meta.LastKnownEnvFingerprint = envFingerprint
	if paths.Layered {
		meta.LastKnownLayeringMode = "layered"
	} else {
		meta.LastKnownLayeringMode = "scoped"
	}
	if cacheHit {
		meta.CacheHitCount++
		meta.ConsecutiveFreshHitCount++
		meta.ConsecutiveMissCount = 0
		meta.LastHitUnixNano = now
	} else {
		meta.CacheMissCount++
		meta.ConsecutiveMissCount++
		meta.ConsecutiveFreshHitCount = 0
		meta.LastLoadUnixNano = now
	}
	meta.Fresh = fresh
	if strings.TrimSpace(lastErr) != "" {
		meta.LastError = strings.TrimSpace(lastErr)
		meta.LastErrorAtUnixNano = now
	}
	meta.LastPersistedAtUnixNano = now
	_ = m.WriteCacheMetadata(projectDir, meta)
}

func (m *Manager) recordInvalidation(projectDir, cause string) {
	meta, err := m.ReadCacheMetadata(projectDir)
	if err != nil {
		return
	}
	meta.Fresh = false
	meta.LastInvalidateCause = cause
	meta.LastInvalidateAt = time.Now().UnixNano()
	meta.InvalidateCount++
	meta.ConsecutiveFreshHitCount = 0
	meta.LastPersistedAtUnixNano = meta.LastInvalidateAt
	_ = m.WriteCacheMetadata(projectDir, meta)
}

func buildEnvFingerprint() string {
	vars := []string{
		"ALLIECODE_DEFAULT_PROVIDER",
		"ALLIECODE_DEFAULT_MODEL",
		"ALLIECODE_PROVIDER_ALLOW",
		"ALLIECODE_MODEL_ALLOW",
		"ALLIECODE_TOOL_PERMISSIONS",
		"ALLIECODE_MAX_RETRIES",
		"ALLIECODE_MAX_TOKENS",
		"ALLIECODE_REQUEST_TIMEOUT_SECONDS",
		"ALLIECODE_REMOTE_MODE",
		"ALLIECODE_REMOTE_LISTEN_ADDR",
		"ALLIECODE_REMOTE_CONNECT_ADDR",
		"ALLIECODE_REMOTE_TOKEN_SOURCE",
		"ALLIECODE_REMOTE_ENABLED",
		"ALLIECODE_COMPANION_NAME",
		"ALLIECODE_COMPANION_PERSONALITY",
		"ALLIECODE_COMPANION_HATCHED_AT",
		"ALLIECODE_COMPANION_MUTED",
		"ALLIECODE_TERMINAL_PROFILE",
		"ALLIECODE_IDE_EDITOR",
		"ALLIECODE_MIGRATION_VERSION",
		"ALLIECODE_RUNTIME_SURFACE",
		"ALLIECODE_RUNTIME_COMMAND_SURFACE",
		"ALLIECODE_STARTUP_HYDRATION_MODE",
		"ALLIECODE_STARTUP_STRICT_HYDRATION",
	}
	sort.Strings(vars)
	b := strings.Builder{}
	for _, k := range vars {
		if v, ok := os.LookupEnv(k); ok {
			b.WriteString(k)
			b.WriteString("=")
			b.WriteString(v)
			b.WriteString("\n")
		}
	}
	h := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(h[:])
}

func signaturesFor(paths []string) ([]fileSignature, error) {
	out := make([]fileSignature, 0, len(paths))
	for _, path := range paths {
		sig, err := signatureFor(path)
		if err != nil {
			return nil, err
		}
		out = append(out, sig)
	}
	return out, nil
}

func signatureFor(path string) (fileSignature, error) {
	clean := filepath.Clean(path)
	st, err := os.Stat(clean)
	if err != nil {
		if os.IsNotExist(err) {
			return fileSignature{path: clean}, nil
		}
		return fileSignature{}, fmt.Errorf("stat %s: %w", clean, err)
	}
	return fileSignature{
		path:    clean,
		exists:  true,
		size:    st.Size(),
		modUnix: st.ModTime().UnixNano(),
	}, nil
}
