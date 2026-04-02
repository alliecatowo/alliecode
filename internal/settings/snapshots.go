package settings

import (
	"sort"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/config"
)

type Snapshot struct {
	Scope            Scope          `json:"scope"`
	ProjectDir       string         `json:"project_dir,omitempty"`
	LoadedAt         time.Time      `json:"loaded_at"`
	DefaultProvider  string         `json:"default_provider,omitempty"`
	DefaultModel     string         `json:"default_model,omitempty"`
	HydrationMode    string         `json:"hydration_mode,omitempty"`
	RuntimeSurface   string         `json:"runtime_surface,omitempty"`
	CommandSurface   string         `json:"command_surface,omitempty"`
	MigrationVersion int            `json:"migration_version,omitempty"`
	ProviderAllow    []string       `json:"provider_allow,omitempty"`
	ModelAllow       []string       `json:"model_allow,omitempty"`
	Raw              *config.Config `json:"-"`
}

type Timeline struct {
	items []Snapshot
}

type SnapshotQuery struct {
	Scope            *Scope
	Provider         string
	Model            string
	RuntimeSurface   string
	HydrationMode    string
	ProjectContains  string
	MinMigration     int
	LoadedAfterUnix  int64
	LoadedBeforeUnix int64
	RequireAllowList bool
	Limit            int
}

func NewTimeline() *Timeline {
	return &Timeline{items: make([]Snapshot, 0, 8)}
}

func SnapshotFromConfig(scope Scope, projectDir string, cfg *config.Config, loadedAt time.Time) Snapshot {
	snap := Snapshot{
		Scope:      scope,
		ProjectDir: strings.TrimSpace(projectDir),
		LoadedAt:   loadedAt.UTC(),
		Raw:        cfg,
	}
	if cfg == nil {
		return snap
	}
	snap.DefaultProvider = strings.TrimSpace(cfg.DefaultProvider)
	snap.DefaultModel = strings.TrimSpace(cfg.DefaultModel)
	snap.HydrationMode = strings.TrimSpace(cfg.Startup.HydrationMode)
	snap.RuntimeSurface = strings.TrimSpace(cfg.Runtime.Surface)
	snap.CommandSurface = strings.TrimSpace(cfg.Runtime.CommandSurface)
	snap.MigrationVersion = cfg.MigrationVersion
	if len(cfg.ProviderAllow) > 0 {
		snap.ProviderAllow = append([]string(nil), cfg.ProviderAllow...)
	}
	if len(cfg.ModelAllow) > 0 {
		snap.ModelAllow = append([]string(nil), cfg.ModelAllow...)
	}
	if snap.HydrationMode == "" {
		snap.HydrationMode = "compat"
	}
	return snap
}

func (t *Timeline) Add(snap Snapshot) {
	if t == nil {
		return
	}
	t.items = append(t.items, snap)
	sort.SliceStable(t.items, func(i, j int) bool {
		return t.items[i].LoadedAt.Before(t.items[j].LoadedAt)
	})
}

func (t *Timeline) Items() []Snapshot {
	if t == nil || len(t.items) == 0 {
		return nil
	}
	out := make([]Snapshot, len(t.items))
	copy(out, t.items)
	return out
}

func (t *Timeline) Query(query SnapshotQuery) []Snapshot {
	if t == nil || len(t.items) == 0 {
		return nil
	}
	provider := strings.ToLower(strings.TrimSpace(query.Provider))
	model := strings.ToLower(strings.TrimSpace(query.Model))
	runtimeSurface := strings.ToLower(strings.TrimSpace(query.RuntimeSurface))
	hydrationMode := strings.ToLower(strings.TrimSpace(query.HydrationMode))
	projectContains := strings.ToLower(strings.TrimSpace(query.ProjectContains))

	out := make([]Snapshot, 0, len(t.items))
	for _, item := range t.items {
		if query.Scope != nil && item.Scope != *query.Scope {
			continue
		}
		if provider != "" && strings.ToLower(item.DefaultProvider) != provider {
			continue
		}
		if model != "" && strings.ToLower(item.DefaultModel) != model {
			continue
		}
		if runtimeSurface != "" && strings.ToLower(item.RuntimeSurface) != runtimeSurface {
			continue
		}
		if hydrationMode != "" && strings.ToLower(item.HydrationMode) != hydrationMode {
			continue
		}
		if projectContains != "" && !strings.Contains(strings.ToLower(item.ProjectDir), projectContains) {
			continue
		}
		if query.MinMigration > 0 && item.MigrationVersion < query.MinMigration {
			continue
		}
		if query.LoadedAfterUnix > 0 && item.LoadedAt.Unix() < query.LoadedAfterUnix {
			continue
		}
		if query.LoadedBeforeUnix > 0 && item.LoadedAt.Unix() > query.LoadedBeforeUnix {
			continue
		}
		if query.RequireAllowList && len(item.ProviderAllow) == 0 && len(item.ModelAllow) == 0 {
			continue
		}
		out = append(out, item)
	}

	if query.Limit > 0 && len(out) > query.Limit {
		out = out[len(out)-query.Limit:]
	}
	return out
}

func LayerDiagnosticsFromSnapshot(snapshot Snapshot) config.LayerDiagnostics {
	if snapshot.Raw == nil {
		return config.LayerDiagnostics{
			EffectivePath:    "",
			LayeringMode:     "unknown",
			DefaultProvider:  snapshot.DefaultProvider,
			DefaultModel:     snapshot.DefaultModel,
			HydrationMode:    snapshot.HydrationMode,
			MigrationVersion: snapshot.MigrationVersion,
			ValidationPassed: true,
		}
	}
	return snapshot.Raw.LayerDiagnostics(snapshot.ProjectDir)
}
