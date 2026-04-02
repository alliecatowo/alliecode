package config

import "strings"

type LayerPrecedenceSnapshot struct {
	LayerName        string `json:"layer_name"`
	Path             string `json:"path,omitempty"`
	DefaultProvider  string `json:"default_provider,omitempty"`
	DefaultModel     string `json:"default_model,omitempty"`
	HydrationMode    string `json:"hydration_mode,omitempty"`
	RuntimeSurface   string `json:"runtime_surface,omitempty"`
	CommandSurface   string `json:"command_surface,omitempty"`
	MigrationVersion int    `json:"migration_version,omitempty"`
	ProviderAllowLen int    `json:"provider_allow_len,omitempty"`
	ModelAllowLen    int    `json:"model_allow_len,omitempty"`
}

type LayerPrecedenceQuery struct {
	LayerNameContains string
	Provider          string
	Model             string
	HydrationMode     string
	RuntimeSurface    string
	Limit             int
}

func BuildLayerPrecedence(projectDir string) ([]LayerPrecedenceSnapshot, error) {
	globalPath, projectPath, err := ConventionalPaths(projectDir)
	if err != nil {
		return nil, err
	}
	globalCfg, err := loadLayer(globalPath)
	if err != nil {
		return nil, err
	}
	projectCfg, err := loadLayer(projectPath)
	if err != nil {
		return nil, err
	}
	envCfg, err := envOverrides()
	if err != nil {
		return nil, err
	}
	effective := Merge(NewDefaultConfig(), globalCfg, projectCfg, envCfg)

	items := []LayerPrecedenceSnapshot{
		snapshotFromConfig("defaults", "", NewDefaultConfig()),
		snapshotFromConfig("global", globalPath, globalCfg),
		snapshotFromConfig("project", projectPath, projectCfg),
		snapshotFromConfig("env", "env", envCfg),
		snapshotFromConfig("effective", projectPath, effective),
	}
	return items, nil
}

func QueryLayerPrecedence(items []LayerPrecedenceSnapshot, query LayerPrecedenceQuery) []LayerPrecedenceSnapshot {
	if len(items) == 0 {
		return nil
	}
	layerContains := strings.ToLower(strings.TrimSpace(query.LayerNameContains))
	provider := strings.ToLower(strings.TrimSpace(query.Provider))
	model := strings.ToLower(strings.TrimSpace(query.Model))
	hydration := strings.ToLower(strings.TrimSpace(query.HydrationMode))
	runtimeSurface := strings.ToLower(strings.TrimSpace(query.RuntimeSurface))

	out := make([]LayerPrecedenceSnapshot, 0, len(items))
	for _, item := range items {
		if layerContains != "" && !strings.Contains(strings.ToLower(item.LayerName), layerContains) {
			continue
		}
		if provider != "" && strings.ToLower(strings.TrimSpace(item.DefaultProvider)) != provider {
			continue
		}
		if model != "" && strings.ToLower(strings.TrimSpace(item.DefaultModel)) != model {
			continue
		}
		if hydration != "" && strings.ToLower(strings.TrimSpace(item.HydrationMode)) != hydration {
			continue
		}
		if runtimeSurface != "" && strings.ToLower(strings.TrimSpace(item.RuntimeSurface)) != runtimeSurface {
			continue
		}
		out = append(out, item)
	}
	if query.Limit > 0 && len(out) > query.Limit {
		out = out[:query.Limit]
	}
	return out
}

func snapshotFromConfig(layerName, path string, cfg *Config) LayerPrecedenceSnapshot {
	out := LayerPrecedenceSnapshot{LayerName: layerName, Path: path}
	if cfg == nil {
		return out
	}
	out.DefaultProvider = strings.TrimSpace(cfg.DefaultProvider)
	out.DefaultModel = strings.TrimSpace(cfg.DefaultModel)
	out.HydrationMode = strings.TrimSpace(cfg.Startup.HydrationMode)
	out.RuntimeSurface = strings.TrimSpace(cfg.Runtime.Surface)
	out.CommandSurface = strings.TrimSpace(cfg.Runtime.CommandSurface)
	out.MigrationVersion = cfg.MigrationVersion
	out.ProviderAllowLen = len(cfg.ProviderAllow)
	out.ModelAllowLen = len(cfg.ModelAllow)
	if out.HydrationMode == "" {
		out.HydrationMode = "compat"
	}
	return out
}
