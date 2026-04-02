package config

import (
	"sort"
	"strings"
)

type LayerRepository struct {
	projectDir string
}

type LayerQuery struct {
	LayeringMode    string
	Provider        string
	Model           string
	RemoteMode      string
	HydrationMode   string
	ValidationState string
	ProviderFromEnv *bool
	ModelFromEnv    *bool
	MinMigration    int
	MaxMigration    int
	Limit           int
}

func NewLayerRepository(projectDir string) *LayerRepository {
	return &LayerRepository{projectDir: strings.TrimSpace(projectDir)}
}

func (r *LayerRepository) Diagnostics() (LayerDiagnostics, error) {
	cfg, err := LoadLayered(r.projectDir)
	if err != nil {
		return LayerDiagnostics{}, err
	}
	return cfg.LayerDiagnostics(r.projectDir), nil
}

func (r *LayerRepository) QueryDiagnosticsPage(query LayerPageQuery) (LayerPage, error) {
	item, err := r.Diagnostics()
	if err != nil {
		return LayerPage{}, err
	}
	return QueryLayerDiagnosticsPage([]LayerDiagnostics{item}, query), nil
}

func (r *LayerRepository) QueryDiagnosticsSummary(query LayerPageQuery) (LayerSummary, error) {
	page, err := r.QueryDiagnosticsPage(query)
	if err != nil {
		return LayerSummary{}, err
	}
	return SummarizeLayerDiagnostics(page.Items), nil
}

func FilterLayerDiagnostics(items []LayerDiagnostics, query LayerQuery) []LayerDiagnostics {
	if len(items) == 0 {
		return nil
	}
	layeringMode := strings.ToLower(strings.TrimSpace(query.LayeringMode))
	provider := strings.ToLower(strings.TrimSpace(query.Provider))
	model := strings.ToLower(strings.TrimSpace(query.Model))
	remoteMode := strings.ToLower(strings.TrimSpace(query.RemoteMode))
	hydrationMode := strings.ToLower(strings.TrimSpace(query.HydrationMode))
	validationState := strings.ToLower(strings.TrimSpace(query.ValidationState))

	out := make([]LayerDiagnostics, 0, len(items))
	for _, item := range items {
		if layeringMode != "" && !strings.Contains(strings.ToLower(strings.TrimSpace(item.LayeringMode)), layeringMode) {
			continue
		}
		if provider != "" && strings.ToLower(strings.TrimSpace(item.DefaultProvider)) != provider {
			continue
		}
		if model != "" && strings.ToLower(strings.TrimSpace(item.DefaultModel)) != model {
			continue
		}
		if remoteMode != "" && strings.ToLower(strings.TrimSpace(item.RemoteMode)) != remoteMode {
			continue
		}
		if hydrationMode != "" && strings.ToLower(strings.TrimSpace(item.HydrationMode)) != hydrationMode {
			continue
		}
		if validationState == "passed" && !item.ValidationPassed {
			continue
		}
		if validationState == "failed" && item.ValidationPassed {
			continue
		}
		if query.ProviderFromEnv != nil && item.ProviderFromEnv != *query.ProviderFromEnv {
			continue
		}
		if query.ModelFromEnv != nil && item.ModelFromEnv != *query.ModelFromEnv {
			continue
		}
		if query.MinMigration > 0 && item.MigrationVersion < query.MinMigration {
			continue
		}
		if query.MaxMigration > 0 && item.MigrationVersion > query.MaxMigration {
			continue
		}
		out = append(out, item)
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].MigrationVersion != out[j].MigrationVersion {
			return out[i].MigrationVersion > out[j].MigrationVersion
		}
		return out[i].EffectivePath < out[j].EffectivePath
	})
	if query.Limit > 0 && len(out) > query.Limit {
		out = out[:query.Limit]
	}
	return out
}
