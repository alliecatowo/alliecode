package plugins

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// PolicyState is the persisted plugin policy model.
// - Installed tracks plugins installed through marketplace flows.
// - Enabled and Disabled track runtime activation preferences.
type PolicyState struct {
	Installed   []string                          `yaml:"installed,omitempty"`
	Enabled     []string                          `yaml:"enabled,omitempty"`
	Disabled    []string                          `yaml:"disabled,omitempty"`
	Pins        map[string]string                 `yaml:"pins,omitempty"`
	Summaries   map[string]PluginStateSummary     `yaml:"summaries,omitempty"`
	Diagnostics map[string][]ValidationDiagnostic `yaml:"diagnostics,omitempty"`
	History     []PluginAction                    `yaml:"history,omitempty"`
}

// RuntimePolicy returns a runtime policy filtered to discovered plugin IDs.
// Unknown IDs in persisted state are ignored so stale policy entries do not
// fail runtime loading when plugins were removed from disk.
func (s PolicyState) RuntimePolicy(discovered []string) Policy {
	known := make(map[string]struct{}, len(discovered))
	for _, id := range discovered {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		known[trimmed] = struct{}{}
	}

	filterKnown := func(ids []string) []string {
		out := make([]string, 0, len(ids))
		seen := make(map[string]struct{}, len(ids))
		for _, id := range ids {
			trimmed := strings.TrimSpace(id)
			if trimmed == "" {
				continue
			}
			if _, ok := known[trimmed]; !ok {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			out = append(out, trimmed)
		}
		sort.Strings(out)
		return out
	}

	return Policy{
		Enabled:  filterKnown(s.Enabled),
		Disabled: filterKnown(s.Disabled),
	}
}

func (s *PolicyState) normalize() {
	normalizeIDs := func(ids []string) []string {
		seen := make(map[string]struct{}, len(ids))
		out := make([]string, 0, len(ids))
		for _, id := range ids {
			trimmed := strings.TrimSpace(id)
			if trimmed == "" {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			out = append(out, trimmed)
		}
		sort.Strings(out)
		return out
	}

	s.Installed = normalizeIDs(s.Installed)
	s.Enabled = normalizeIDs(s.Enabled)
	s.Disabled = normalizeIDs(s.Disabled)

	if len(s.Pins) > 0 {
		normalizedPins := make(map[string]string, len(s.Pins))
		for id, version := range s.Pins {
			trimmedID := strings.TrimSpace(id)
			trimmedVersion := strings.TrimSpace(version)
			if trimmedID == "" || trimmedVersion == "" {
				continue
			}
			normalizedPins[trimmedID] = trimmedVersion
		}
		if len(normalizedPins) == 0 {
			s.Pins = nil
		} else {
			s.Pins = normalizedPins
		}
	}

	disabledSet := make(map[string]struct{}, len(s.Disabled))
	for _, id := range s.Disabled {
		disabledSet[id] = struct{}{}
	}
	filteredEnabled := make([]string, 0, len(s.Enabled))
	for _, id := range s.Enabled {
		if _, disabled := disabledSet[id]; disabled {
			continue
		}
		filteredEnabled = append(filteredEnabled, id)
	}
	s.Enabled = filteredEnabled

	if len(s.Summaries) > 0 {
		normalizedSummaries := make(map[string]PluginStateSummary, len(s.Summaries))
		for id, summary := range s.Summaries {
			trimmedID := strings.TrimSpace(id)
			if trimmedID == "" {
				continue
			}
			summary.PluginID = trimmedID
			summary.Version = strings.TrimSpace(summary.Version)
			summary.PinnedVersion = strings.TrimSpace(summary.PinnedVersion)
			summary.PolicyStatus = strings.TrimSpace(summary.PolicyStatus)
			summary.LastAction = strings.TrimSpace(summary.LastAction)
			normalizedSummaries[trimmedID] = summary
		}
		if len(normalizedSummaries) == 0 {
			s.Summaries = nil
		} else {
			s.Summaries = normalizedSummaries
		}
	}

	if len(s.Diagnostics) > 0 {
		normalizedDiagnostics := make(map[string][]ValidationDiagnostic, len(s.Diagnostics))
		for id, diags := range s.Diagnostics {
			trimmedID := strings.TrimSpace(id)
			if trimmedID == "" {
				continue
			}
			normalized := make([]ValidationDiagnostic, 0, len(diags))
			seen := make(map[string]struct{}, len(diags))
			for _, diag := range diags {
				diag.PluginID = trimmedID
				diag.Severity = strings.ToLower(strings.TrimSpace(diag.Severity))
				if diag.Severity == "" {
					diag.Severity = "error"
				}
				diag.Code = strings.TrimSpace(diag.Code)
				diag.Message = strings.TrimSpace(diag.Message)
				if diag.Code == "" || diag.Message == "" {
					continue
				}
				key := diag.Severity + "\n" + diag.Code + "\n" + diag.Message
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				normalized = append(normalized, diag)
			}
			sort.Slice(normalized, func(i, j int) bool {
				if normalized[i].Severity != normalized[j].Severity {
					return normalized[i].Severity < normalized[j].Severity
				}
				if normalized[i].Code != normalized[j].Code {
					return normalized[i].Code < normalized[j].Code
				}
				return normalized[i].Message < normalized[j].Message
			})
			if len(normalized) > 0 {
				normalizedDiagnostics[trimmedID] = normalized
			}
		}
		if len(normalizedDiagnostics) == 0 {
			s.Diagnostics = nil
		} else {
			s.Diagnostics = normalizedDiagnostics
		}
	}

	if len(s.History) > 0 {
		normalizedHistory := make([]PluginAction, 0, len(s.History))
		for _, event := range s.History {
			event.Timestamp = strings.TrimSpace(event.Timestamp)
			event.Action = strings.TrimSpace(strings.ToLower(event.Action))
			event.PluginID = strings.TrimSpace(event.PluginID)
			event.RequestedVersion = strings.TrimSpace(event.RequestedVersion)
			event.PreviousVersion = strings.TrimSpace(event.PreviousVersion)
			event.Version = strings.TrimSpace(event.Version)
			event.Status = strings.TrimSpace(strings.ToLower(event.Status))
			event.Detail = strings.TrimSpace(event.Detail)
			if event.Action == "" && event.PluginID == "" {
				continue
			}
			normalizedHistory = append(normalizedHistory, event)
		}
		if len(normalizedHistory) == 0 {
			s.History = nil
		} else {
			s.History = normalizedHistory
		}
	}
}

// PolicyStore persists plugin policy state to disk.
type PolicyStore struct {
	path string
}

// NewPolicyStore creates a new store bound to one policy file path.
func NewPolicyStore(path string) *PolicyStore {
	return &PolicyStore{path: path}
}

// Load reads policy state from disk. Missing files return an empty state.
func (s *PolicyStore) Load() (PolicyState, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return PolicyState{}, nil
		}
		return PolicyState{}, fmt.Errorf("read policy store %q: %w", s.path, err)
	}

	var state PolicyState
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&state); err != nil {
		return PolicyState{}, fmt.Errorf("decode policy store %q: %w", s.path, err)
	}
	state.normalize()
	return state, nil
}

// Save writes policy state atomically to disk.
func (s *PolicyStore) Save(state PolicyState) error {
	state.normalize()

	encoded, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("encode policy store %q: %w", s.path, err)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create policy store dir for %q: %w", s.path, err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(s.path), ".policy-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary policy file for %q: %w", s.path, err)
	}
	tmpPath := tmp.Name()

	writeErr := func(err error) error {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return err
	}

	if _, err := tmp.Write(encoded); err != nil {
		return writeErr(fmt.Errorf("write temporary policy file %q: %w", tmpPath, err))
	}
	if err := tmp.Chmod(0o644); err != nil {
		return writeErr(fmt.Errorf("chmod temporary policy file %q: %w", tmpPath, err))
	}
	if err := tmp.Close(); err != nil {
		return writeErr(fmt.Errorf("close temporary policy file %q: %w", tmpPath, err))
	}
	if err := os.Rename(tmpPath, s.path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("replace policy store %q: %w", s.path, err)
	}

	return nil
}
