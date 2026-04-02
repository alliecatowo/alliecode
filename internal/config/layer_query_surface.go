package config

import "strings"

type LayerPageQuery struct {
	LayerQuery
	Offset                int
	Limit                 int
	PathContains          string
	ValidationErrContains string
	RemoteMode            string
	RequireValidationErr  *bool
}

type LayerPage struct {
	Items   []LayerDiagnostics `json:"items,omitempty"`
	Total   int                `json:"total"`
	Offset  int                `json:"offset"`
	Limit   int                `json:"limit"`
	HasMore bool               `json:"has_more"`
}

type LayerSummary struct {
	Total            int            `json:"total"`
	ValidationPassed int            `json:"validation_passed"`
	ValidationFailed int            `json:"validation_failed"`
	StrictHydration  int            `json:"strict_hydration"`
	LayeringModes    map[string]int `json:"layering_modes,omitempty"`
	HydrationModes   map[string]int `json:"hydration_modes,omitempty"`
	RemoteModes      map[string]int `json:"remote_modes,omitempty"`
}

func QueryLayerDiagnosticsPage(items []LayerDiagnostics, query LayerPageQuery) LayerPage {
	base := query.LayerQuery
	base.Limit = 0
	filtered := FilterLayerDiagnostics(items, base)
	filtered = filterLayerSurface(filtered, query)
	total := len(filtered)
	start, end := diagPagination(total, query.Offset, query.Limit)
	return LayerPage{Items: filtered[start:end], Total: total, Offset: start, Limit: end - start, HasMore: end < total}
}

func SummarizeLayerDiagnostics(items []LayerDiagnostics) LayerSummary {
	out := LayerSummary{LayeringModes: make(map[string]int), HydrationModes: make(map[string]int), RemoteModes: make(map[string]int)}
	for _, item := range items {
		out.Total++
		if item.ValidationPassed {
			out.ValidationPassed++
		} else {
			out.ValidationFailed++
		}
		if v := strings.TrimSpace(item.LayeringMode); v != "" {
			out.LayeringModes[v]++
		}
		if v := strings.TrimSpace(item.HydrationMode); v != "" {
			out.HydrationModes[v]++
		}
		if item.StrictHydration {
			out.StrictHydration++
		}
		if v := strings.TrimSpace(item.RemoteMode); v != "" {
			out.RemoteModes[v]++
		}
	}
	if len(out.LayeringModes) == 0 {
		out.LayeringModes = nil
	}
	if len(out.HydrationModes) == 0 {
		out.HydrationModes = nil
	}
	if len(out.RemoteModes) == 0 {
		out.RemoteModes = nil
	}
	return out
}

func filterLayerSurface(items []LayerDiagnostics, query LayerPageQuery) []LayerDiagnostics {
	if len(items) == 0 {
		return nil
	}
	pathContains := strings.ToLower(strings.TrimSpace(query.PathContains))
	validationErrContains := strings.ToLower(strings.TrimSpace(query.ValidationErrContains))
	remoteMode := strings.ToLower(strings.TrimSpace(query.RemoteMode))
	out := make([]LayerDiagnostics, 0, len(items))
	for _, item := range items {
		if pathContains != "" && !strings.Contains(strings.ToLower(strings.TrimSpace(item.EffectivePath)), pathContains) {
			continue
		}
		if validationErrContains != "" && !strings.Contains(strings.ToLower(strings.TrimSpace(item.ValidationError)), validationErrContains) {
			continue
		}
		if remoteMode != "" && strings.ToLower(strings.TrimSpace(item.RemoteMode)) != remoteMode {
			continue
		}
		if query.RequireValidationErr != nil {
			hasErr := strings.TrimSpace(item.ValidationError) != ""
			if hasErr != *query.RequireValidationErr {
				continue
			}
		}
		out = append(out, item)
	}
	return out
}
