package settings

import (
	"sort"

	"github.com/alliecatowo/alliecode/internal/config"
)

type ValidationDiagnostic struct {
	Scope   Scope  `json:"scope"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func ValidationDiagnostics(scope Scope, cfg *config.Config) []ValidationDiagnostic {
	if cfg == nil {
		return []ValidationDiagnostic{{
			Scope:   scope,
			Code:    "config_nil",
			Message: "configuration is nil",
		}}
	}

	err := cfg.Validate()
	if err == nil {
		return nil
	}

	parts := flattenErrors(err)
	out := make([]ValidationDiagnostic, 0, len(parts))
	for _, part := range parts {
		out = append(out, ValidationDiagnostic{
			Scope:   scope,
			Code:    "config_invalid",
			Message: part,
		})
	}
	return out
}

func AggregateValidationDiagnostics(items map[Scope]*config.Config) []ValidationDiagnostic {
	if len(items) == 0 {
		return nil
	}

	scopes := make([]Scope, 0, len(items))
	for scope := range items {
		scopes = append(scopes, scope)
	}
	sort.Slice(scopes, func(i, j int) bool { return scopes[i] < scopes[j] })

	out := make([]ValidationDiagnostic, 0)
	for _, scope := range scopes {
		out = append(out, ValidationDiagnostics(scope, items[scope])...)
	}
	return out
}

func flattenErrors(err error) []string {
	if err == nil {
		return nil
	}

	type unwrapMany interface{ Unwrap() []error }
	var out []string
	var walk func(error)
	walk = func(current error) {
		if current == nil {
			return
		}
		if many, ok := current.(unwrapMany); ok {
			for _, nested := range many.Unwrap() {
				walk(nested)
			}
			return
		}
		out = append(out, current.Error())
	}

	walk(err)
	if len(out) == 0 {
		return []string{err.Error()}
	}
	return out
}
