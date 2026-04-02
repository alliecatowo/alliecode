package config

import (
	"sort"
	"strings"
)

type MigrationReport struct {
	LayeringMode      string   `json:"layering_mode,omitempty"`
	HydrationMode     string   `json:"hydration_mode,omitempty"`
	StrictHydration   bool     `json:"strict_hydration"`
	RemoteMode        string   `json:"remote_mode,omitempty"`
	DefaultProvider   string   `json:"default_provider,omitempty"`
	DefaultModel      string   `json:"default_model,omitempty"`
	MigrationVersion  int      `json:"migration_version,omitempty"`
	AppliedPatchCount int      `json:"applied_patch_count,omitempty"`
	AppliedPatches    []string `json:"applied_patches,omitempty"`
	ValidationPassed  bool     `json:"validation_passed"`
	Summary           string   `json:"summary,omitempty"`
}

type MigrationReportQuery struct {
	LayeringMode     string
	HydrationMode    string
	RemoteMode       string
	Provider         string
	Model            string
	StrictHydration  *bool
	ValidationPassed *bool
	MinVersion       int
	MinApplied       int
	SummaryContains  string
	Limit            int
}

func BuildMigrationReport(cfg *Config, projectDir string, appliedPatches []string) MigrationReport {
	if cfg == nil {
		return MigrationReport{}
	}
	diag := cfg.LayerDiagnostics(projectDir)
	summaryParts := []string{diag.LayeringMode, diag.HydrationMode, diag.RemoteMode, diag.DefaultProvider, diag.DefaultModel}
	return MigrationReport{
		LayeringMode:      diag.LayeringMode,
		HydrationMode:     diag.HydrationMode,
		StrictHydration:   diag.StrictHydration,
		RemoteMode:        diag.RemoteMode,
		DefaultProvider:   diag.DefaultProvider,
		DefaultModel:      diag.DefaultModel,
		MigrationVersion:  cfg.MigrationVersion,
		AppliedPatchCount: len(appliedPatches),
		AppliedPatches:    append([]string(nil), appliedPatches...),
		ValidationPassed:  diag.ValidationPassed,
		Summary:           strings.TrimSpace(strings.Join(summaryParts, "|")),
	}
}

func QueryMigrationReports(items []MigrationReport, query MigrationReportQuery) []MigrationReport {
	if len(items) == 0 {
		return nil
	}
	layeringMode := strings.ToLower(strings.TrimSpace(query.LayeringMode))
	hydrationMode := strings.ToLower(strings.TrimSpace(query.HydrationMode))
	remoteMode := strings.ToLower(strings.TrimSpace(query.RemoteMode))
	provider := strings.ToLower(strings.TrimSpace(query.Provider))
	model := strings.ToLower(strings.TrimSpace(query.Model))
	summaryContains := strings.ToLower(strings.TrimSpace(query.SummaryContains))

	out := make([]MigrationReport, 0, len(items))
	for _, item := range items {
		if layeringMode != "" && strings.ToLower(strings.TrimSpace(item.LayeringMode)) != layeringMode {
			continue
		}
		if hydrationMode != "" && strings.ToLower(strings.TrimSpace(item.HydrationMode)) != hydrationMode {
			continue
		}
		if remoteMode != "" && strings.ToLower(strings.TrimSpace(item.RemoteMode)) != remoteMode {
			continue
		}
		if provider != "" && strings.ToLower(strings.TrimSpace(item.DefaultProvider)) != provider {
			continue
		}
		if model != "" && strings.ToLower(strings.TrimSpace(item.DefaultModel)) != model {
			continue
		}
		if query.StrictHydration != nil && item.StrictHydration != *query.StrictHydration {
			continue
		}
		if query.ValidationPassed != nil && item.ValidationPassed != *query.ValidationPassed {
			continue
		}
		if query.MinVersion > 0 && item.MigrationVersion < query.MinVersion {
			continue
		}
		if query.MinApplied > 0 && item.AppliedPatchCount < query.MinApplied {
			continue
		}
		if summaryContains != "" && !strings.Contains(strings.ToLower(item.Summary), summaryContains) {
			continue
		}
		out = append(out, item)
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].MigrationVersion != out[j].MigrationVersion {
			return out[i].MigrationVersion > out[j].MigrationVersion
		}
		if out[i].AppliedPatchCount != out[j].AppliedPatchCount {
			return out[i].AppliedPatchCount > out[j].AppliedPatchCount
		}
		return out[i].LayeringMode < out[j].LayeringMode
	})
	if query.Limit > 0 && len(out) > query.Limit {
		out = out[:query.Limit]
	}
	return out
}
