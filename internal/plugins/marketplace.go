package plugins

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// RegistryPlugin is one plugin entry in a local marketplace index file.
type RegistryPlugin struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name,omitempty"`
	Version     string `yaml:"version"`
	SourcePath  string `yaml:"source_path"`
	Description string `yaml:"description,omitempty"`
}

// RegistryIndex is the marketplace-style plugin index representation.
type RegistryIndex struct {
	Plugins []RegistryPlugin `yaml:"plugins"`
}

// Marketplace resolves installable plugins from a local index file.
type Marketplace struct {
	IndexPath string
}

// LoadIndex reads and validates the local marketplace index.
func (m *Marketplace) LoadIndex() (RegistryIndex, error) {
	data, err := os.ReadFile(m.IndexPath)
	if err != nil {
		return RegistryIndex{}, fmt.Errorf("read marketplace index %q: %w", m.IndexPath, err)
	}

	var index RegistryIndex
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&index); err != nil {
		return RegistryIndex{}, fmt.Errorf("decode marketplace index %q: %w", m.IndexPath, err)
	}

	seen := make(map[string]struct{}, len(index.Plugins))
	for i := range index.Plugins {
		entry := &index.Plugins[i]
		entry.ID = strings.TrimSpace(entry.ID)
		entry.Name = strings.TrimSpace(entry.Name)
		entry.Version = strings.TrimSpace(entry.Version)
		entry.SourcePath = strings.TrimSpace(entry.SourcePath)
		entry.Description = strings.TrimSpace(entry.Description)

		if entry.ID == "" {
			return RegistryIndex{}, fmt.Errorf("plugins[%d].id is required", i)
		}
		if !pluginIDPattern.MatchString(entry.ID) {
			return RegistryIndex{}, fmt.Errorf("plugins[%d].id %q must match %s", i, entry.ID, pluginIDPattern.String())
		}
		if entry.Version == "" {
			return RegistryIndex{}, fmt.Errorf("plugins[%d].version is required", i)
		}
		if entry.SourcePath == "" {
			return RegistryIndex{}, fmt.Errorf("plugins[%d].source_path is required", i)
		}
		key := entry.ID + "@" + entry.Version
		if _, exists := seen[key]; exists {
			return RegistryIndex{}, fmt.Errorf("duplicate plugin id/version %q in marketplace index", key)
		}
		seen[key] = struct{}{}
	}

	return index, nil
}

// Resolve returns one plugin entry by ID.
func (m *Marketplace) Resolve(pluginID string) (RegistryPlugin, error) {
	index, err := m.LoadIndex()
	if err != nil {
		return RegistryPlugin{}, err
	}
	targetID, pinnedVersion := splitPluginRef(pluginID)
	if targetID == "" {
		return RegistryPlugin{}, fmt.Errorf("plugin id is required")
	}

	matches := make([]RegistryPlugin, 0, len(index.Plugins))
	for _, plugin := range index.Plugins {
		if plugin.ID != targetID {
			continue
		}
		if pinnedVersion != "" && plugin.Version != pinnedVersion {
			continue
		}
		matches = append(matches, plugin)
	}

	if len(matches) == 0 {
		if pinnedVersion != "" {
			return RegistryPlugin{}, fmt.Errorf("plugin %q with version %q not found in marketplace index", targetID, pinnedVersion)
		}
		return RegistryPlugin{}, fmt.Errorf("plugin %q not found in marketplace index", targetID)
	}

	sort.Slice(matches, func(i, j int) bool {
		cmp := compareVersions(matches[i].Version, matches[j].Version)
		if cmp != 0 {
			return cmp > 0
		}
		return matches[i].SourcePath < matches[j].SourcePath
	})
	return matches[0], nil
}

// Service coordinates install/list/enable/disable policy flows.
type Service struct {
	InstallRoot string
	Store       *PolicyStore
	Marketplace *Marketplace
}

// PluginAction is one persisted plugin lifecycle event for auditability.
type PluginAction struct {
	Timestamp        string `yaml:"timestamp,omitempty" json:"timestamp,omitempty"`
	Action           string `yaml:"action,omitempty" json:"action,omitempty"`
	PluginID         string `yaml:"plugin_id,omitempty" json:"plugin_id,omitempty"`
	RequestedVersion string `yaml:"requested_version,omitempty" json:"requested_version,omitempty"`
	PreviousVersion  string `yaml:"previous_version,omitempty" json:"previous_version,omitempty"`
	Version          string `yaml:"version,omitempty" json:"version,omitempty"`
	Status           string `yaml:"status,omitempty" json:"status,omitempty"`
	Detail           string `yaml:"detail,omitempty" json:"detail,omitempty"`
}

// ValidationDiagnostic captures deterministic validation issues per plugin.
type ValidationDiagnostic struct {
	PluginID string `yaml:"plugin_id,omitempty" json:"plugin_id"`
	Severity string `yaml:"severity,omitempty" json:"severity"`
	Code     string `yaml:"code,omitempty" json:"code"`
	Message  string `yaml:"message,omitempty" json:"message"`
}

// PluginStateSummary is persisted state for one plugin across operations.
type PluginStateSummary struct {
	PluginID         string `yaml:"plugin_id,omitempty" json:"plugin_id"`
	Version          string `yaml:"version,omitempty" json:"version,omitempty"`
	PinnedVersion    string `yaml:"pinned_version,omitempty" json:"pinned_version,omitempty"`
	PolicyStatus     string `yaml:"policy_status,omitempty" json:"policy_status,omitempty"`
	Installed        bool   `yaml:"installed,omitempty" json:"installed"`
	Enabled          bool   `yaml:"enabled,omitempty" json:"enabled"`
	LastAction       string `yaml:"last_action,omitempty" json:"last_action,omitempty"`
	LastUpdatedAt    string `yaml:"last_updated_at,omitempty" json:"last_updated_at,omitempty"`
	DiagnosticsCount int    `yaml:"diagnostics_count,omitempty" json:"diagnostics_count"`
}

// PluginPolicySummary describes current policy intent and effective state.
type PluginPolicySummary struct {
	PluginID         string `json:"plugin_id"`
	PolicyInstalled  bool   `json:"policy_installed"`
	PolicyEnabled    bool   `json:"policy_enabled"`
	PolicyDisabled   bool   `json:"policy_disabled"`
	PinnedVersion    string `json:"pinned_version,omitempty"`
	InstalledVersion string `json:"installed_version,omitempty"`
	EffectiveEnabled bool   `json:"effective_enabled"`
	HasDiagnostics   bool   `json:"has_diagnostics"`
	DiagnosticsCount int    `json:"diagnostics_count"`
}

// ServiceListResult contains plugin inventory plus validation diagnostics.
type ServiceListResult struct {
	Plugins       []InstalledPlugin                 `json:"plugins"`
	Summaries     map[string]PluginStateSummary     `json:"summaries,omitempty"`
	PolicySummary map[string]PluginPolicySummary    `json:"policy_summary,omitempty"`
	Diagnostics   map[string][]ValidationDiagnostic `json:"diagnostics,omitempty"`
	History       []PluginAction                    `json:"history,omitempty"`
}

// NewService builds a plugin service.
func NewService(installRoot, policyPath, indexPath string) *Service {
	return &Service{
		InstallRoot: installRoot,
		Store:       NewPolicyStore(policyPath),
		Marketplace: &Marketplace{IndexPath: indexPath},
	}
}

// Install installs a plugin from the local marketplace index and enables it.
// pluginRef accepts either "plugin-id" or "plugin-id@version".
func (s *Service) Install(pluginRef string) error {
	id, requestedVersion := splitPluginRef(pluginRef)
	entry, err := s.Marketplace.Resolve(pluginRef)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(s.InstallRoot, 0o755); err != nil {
		return fmt.Errorf("create install root %q: %w", s.InstallRoot, err)
	}

	targetDir := filepath.Join(s.InstallRoot, id)
	if st, err := os.Stat(targetDir); err == nil && st.IsDir() {
		return fmt.Errorf("plugin %q is already installed (run update to change version)", id)
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("check install target %q: %w", targetDir, err)
	}

	sourceDir := entry.SourcePath
	if !filepath.IsAbs(sourceDir) {
		sourceDir = filepath.Join(filepath.Dir(s.Marketplace.IndexPath), sourceDir)
	}
	if err := copyDir(sourceDir, targetDir); err != nil {
		return fmt.Errorf("install plugin %q from %q: %w", id, sourceDir, err)
	}

	if _, err := LoadRuntime(s.InstallRoot, Policy{}); err != nil {
		_ = os.RemoveAll(targetDir)
		return fmt.Errorf("validate installed plugins after installing %q: %w", id, err)
	}

	state, err := s.Store.Load()
	if err != nil {
		return err
	}
	state.Installed = append(state.Installed, id)
	state.Enabled = append(state.Enabled, id)
	state.Disabled = removeID(state.Disabled, id)
	if requestedVersion != "" {
		if state.Pins == nil {
			state.Pins = map[string]string{}
		}
		state.Pins[id] = requestedVersion
	}

	if err := s.Store.Save(state); err != nil {
		return err
	}
	if err := s.refreshPersistedDiagnosticsAndSummaries(id, "install"); err != nil {
		return err
	}
	if err := s.recordAction(PluginAction{
		Action:           "install",
		PluginID:         id,
		RequestedVersion: requestedVersion,
		Version:          entry.Version,
		Status:           "ok",
	}); err != nil {
		return err
	}
	return nil
}

// Update updates one installed plugin from marketplace index.
// pluginRef accepts either "plugin-id" or "plugin-id@version".
func (s *Service) Update(pluginRef string) error {
	id, requestedVersion := splitPluginRef(pluginRef)
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("plugin id is required")
	}

	known, err := s.discoveredPluginIDs()
	if err != nil {
		return err
	}
	if _, ok := known[id]; !ok {
		return fmt.Errorf("plugin %q is not installed", id)
	}

	state, err := s.Store.Load()
	if err != nil {
		return err
	}

	resolvedRef := pluginRef
	if requestedVersion == "" {
		if pinnedVersion := strings.TrimSpace(state.Pins[id]); pinnedVersion != "" {
			resolvedRef = id + "@" + pinnedVersion
			requestedVersion = pinnedVersion
		}
	}

	entry, err := s.Marketplace.Resolve(resolvedRef)
	if err != nil {
		return err
	}

	targetDir := filepath.Join(s.InstallRoot, id)
	previousVersion, err := installedPluginVersion(targetDir)
	if err != nil {
		return err
	}
	if previousVersion == entry.Version {
		if requestedVersion != "" {
			if state.Pins == nil {
				state.Pins = map[string]string{}
			}
			state.Pins[id] = requestedVersion
			if err := s.Store.Save(state); err != nil {
				return err
			}
		}
		if err := s.refreshPersistedDiagnosticsAndSummaries(id, "update"); err != nil {
			return err
		}
		return s.recordAction(PluginAction{
			Action:           "update",
			PluginID:         id,
			RequestedVersion: requestedVersion,
			PreviousVersion:  previousVersion,
			Version:          entry.Version,
			Status:           "noop",
			Detail:           "already at requested version",
		})
	}

	sourceDir := entry.SourcePath
	if !filepath.IsAbs(sourceDir) {
		sourceDir = filepath.Join(filepath.Dir(s.Marketplace.IndexPath), sourceDir)
	}

	backupRoot := filepath.Join(filepath.Dir(s.InstallRoot), ".plugin-backups")
	if err := os.MkdirAll(backupRoot, 0o755); err != nil {
		return fmt.Errorf("create plugin backup root %q: %w", backupRoot, err)
	}
	backupDir := filepath.Join(backupRoot, id+"-"+strconv.FormatInt(time.Now().UTC().UnixNano(), 10))
	if err := os.Rename(targetDir, backupDir); err != nil {
		return fmt.Errorf("prepare plugin backup for %q: %w", id, err)
	}
	restored := false
	restore := func() {
		if restored {
			return
		}
		_ = os.RemoveAll(targetDir)
		_ = os.Rename(backupDir, targetDir)
		restored = true
	}

	if err := copyDir(sourceDir, targetDir); err != nil {
		restore()
		return fmt.Errorf("update plugin %q from %q: %w", id, sourceDir, err)
	}
	if _, err := LoadRuntime(s.InstallRoot, Policy{}); err != nil {
		restore()
		return fmt.Errorf("validate installed plugins after updating %q: %w", id, err)
	}
	if !restored {
		_ = os.RemoveAll(backupDir)
	}

	if requestedVersion != "" {
		if state.Pins == nil {
			state.Pins = map[string]string{}
		}
		state.Pins[id] = requestedVersion
		if err := s.Store.Save(state); err != nil {
			return err
		}
	}

	if err := s.refreshPersistedDiagnosticsAndSummaries(id, "update"); err != nil {
		return err
	}
	return s.recordAction(PluginAction{
		Action:           "update",
		PluginID:         id,
		RequestedVersion: requestedVersion,
		PreviousVersion:  previousVersion,
		Version:          entry.Version,
		Status:           "ok",
	})
}

// Remove uninstalls a plugin from the local install root and policy state.
func (s *Service) Remove(pluginRef string) error {
	id, _ := splitPluginRef(pluginRef)
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("plugin id is required")
	}

	targetDir := filepath.Join(s.InstallRoot, id)
	if _, err := os.Stat(targetDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("plugin %q is not installed", id)
		}
		return fmt.Errorf("check install target %q: %w", targetDir, err)
	}

	previousVersion, _ := installedPluginVersion(targetDir)
	if err := os.RemoveAll(targetDir); err != nil {
		return fmt.Errorf("remove installed plugin %q: %w", id, err)
	}

	state, err := s.Store.Load()
	if err != nil {
		return err
	}
	state.Installed = removeID(state.Installed, id)
	state.Enabled = removeID(state.Enabled, id)
	state.Disabled = removeID(state.Disabled, id)
	if state.Pins != nil {
		delete(state.Pins, id)
	}

	if err := s.Store.Save(state); err != nil {
		return err
	}
	if err := s.refreshPersistedDiagnosticsAndSummaries(id, "remove"); err != nil {
		return err
	}
	return s.recordAction(PluginAction{
		Action:          "remove",
		PluginID:        id,
		PreviousVersion: previousVersion,
		Status:          "ok",
	})
}

// Enable marks one installed plugin as enabled in persisted policy.
func (s *Service) Enable(pluginID string) error {
	known, err := s.discoveredPluginIDs()
	if err != nil {
		return err
	}
	id := strings.TrimSpace(pluginID)
	if _, ok := known[id]; !ok {
		return fmt.Errorf("plugin %q is not installed", id)
	}

	state, err := s.Store.Load()
	if err != nil {
		return err
	}
	state.Enabled = append(state.Enabled, id)
	state.Disabled = removeID(state.Disabled, id)
	state.Installed = append(state.Installed, id)
	if err := s.Store.Save(state); err != nil {
		return err
	}
	if err := s.refreshPersistedDiagnosticsAndSummaries(id, "enable"); err != nil {
		return err
	}
	return s.recordAction(PluginAction{Action: "enable", PluginID: id, Status: "ok"})
}

// Disable marks one installed plugin as disabled in persisted policy.
func (s *Service) Disable(pluginID string) error {
	known, err := s.discoveredPluginIDs()
	if err != nil {
		return err
	}
	id := strings.TrimSpace(pluginID)
	if _, ok := known[id]; !ok {
		return fmt.Errorf("plugin %q is not installed", id)
	}

	state, err := s.Store.Load()
	if err != nil {
		return err
	}
	state.Disabled = append(state.Disabled, id)
	state.Enabled = removeID(state.Enabled, id)
	state.Installed = append(state.Installed, id)
	if err := s.Store.Save(state); err != nil {
		return err
	}
	if err := s.refreshPersistedDiagnosticsAndSummaries(id, "disable"); err != nil {
		return err
	}
	return s.recordAction(PluginAction{Action: "disable", PluginID: id, Status: "ok"})
}

// PinVersion pins one installed plugin to an explicit version.
func (s *Service) PinVersion(pluginRef string) error {
	id, version := splitPluginRef(pluginRef)
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("plugin id is required")
	}
	if strings.TrimSpace(version) == "" {
		return fmt.Errorf("version is required for pinning")
	}

	known, err := s.discoveredPluginIDs()
	if err != nil {
		return err
	}
	if _, ok := known[id]; !ok {
		return fmt.Errorf("plugin %q is not installed", id)
	}

	if _, err := s.Marketplace.Resolve(id + "@" + version); err != nil {
		return err
	}

	state, err := s.Store.Load()
	if err != nil {
		return err
	}
	if state.Pins == nil {
		state.Pins = map[string]string{}
	}
	state.Pins[id] = version
	if err := s.Store.Save(state); err != nil {
		return err
	}
	if err := s.refreshPersistedDiagnosticsAndSummaries(id, "pin"); err != nil {
		return err
	}
	return s.recordAction(PluginAction{Action: "pin", PluginID: id, RequestedVersion: version, Version: version, Status: "ok"})
}

// UnpinVersion removes a pin for one installed plugin.
func (s *Service) UnpinVersion(pluginID string) error {
	id := strings.TrimSpace(pluginID)
	if id == "" {
		return fmt.Errorf("plugin id is required")
	}

	state, err := s.Store.Load()
	if err != nil {
		return err
	}
	if state.Pins != nil {
		delete(state.Pins, id)
	}
	if err := s.Store.Save(state); err != nil {
		return err
	}
	if err := s.refreshPersistedDiagnosticsAndSummaries(id, "unpin"); err != nil {
		return err
	}
	return s.recordAction(PluginAction{Action: "unpin", PluginID: id, Status: "ok"})
}

// List returns discovered plugins with current enabled state applied.
func (s *Service) List() ([]InstalledPlugin, error) {
	result, err := s.ListWithDiagnostics()
	if err != nil {
		return nil, err
	}
	return result.Plugins, nil
}

// ListWithDiagnostics returns discovered plugins plus persisted diagnostics and summaries.
func (s *Service) ListWithDiagnostics() (ServiceListResult, error) {
	state, err := s.Store.Load()
	if err != nil {
		return ServiceListResult{}, err
	}

	plugins := make([]InstalledPlugin, 0)
	ids := make([]string, 0)
	var runtimeErr error

	base, err := LoadRuntime(s.InstallRoot, Policy{})
	if err != nil {
		if !os.IsNotExist(err) {
			runtimeErr = err
		}
	} else {
		ids = make([]string, 0, len(base.plugins))
		for _, p := range base.plugins {
			ids = append(ids, p.Manifest.ID)
		}

		runtime, runtimeLoadErr := LoadRuntime(s.InstallRoot, state.RuntimePolicy(ids))
		if runtimeLoadErr != nil {
			runtimeErr = runtimeLoadErr
		} else {
			plugins = runtime.Plugins()
		}
	}

	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Manifest.ID < plugins[j].Manifest.ID
	})

	diagnostics := s.collectValidationDiagnostics(state, runtimeErr)
	summaries := mergePluginStateSummaries(state.Summaries, plugins, diagnostics)
	summaries = applyPolicyFieldsToSummaries(summaries, state)
	policySummary := buildPluginPolicySummary(state, plugins, diagnostics)

	state.Diagnostics = diagnostics
	state.Summaries = summaries
	if err := s.Store.Save(state); err != nil {
		return ServiceListResult{}, err
	}

	return ServiceListResult{Plugins: plugins, Summaries: summaries, PolicySummary: policySummary, Diagnostics: diagnostics, History: append([]PluginAction(nil), state.History...)}, nil
}

func (s *Service) refreshPersistedDiagnosticsAndSummaries(targetPluginID, action string) error {
	state, err := s.Store.Load()
	if err != nil {
		return err
	}

	result, err := s.ListWithDiagnostics()
	if err != nil {
		return err
	}

	if state.Summaries == nil {
		state.Summaries = map[string]PluginStateSummary{}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	target := strings.TrimSpace(targetPluginID)
	if target != "" {
		summary := state.Summaries[target]
		summary.PluginID = target
		summary.LastAction = strings.TrimSpace(action)
		summary.LastUpdatedAt = now
		if discovered, ok := result.Summaries[target]; ok {
			summary.Version = discovered.Version
			summary.PinnedVersion = discovered.PinnedVersion
			summary.PolicyStatus = discovered.PolicyStatus
			summary.Installed = discovered.Installed
			summary.Enabled = discovered.Enabled
			summary.DiagnosticsCount = discovered.DiagnosticsCount
		} else if strings.TrimSpace(action) == "remove" {
			summary.Installed = false
			summary.Enabled = false
		}
		state.Summaries[target] = summary
	}

	state.Diagnostics = result.Diagnostics
	for id, discovered := range result.Summaries {
		prior := state.Summaries[id]
		if prior.LastUpdatedAt == "" {
			prior.LastUpdatedAt = now
		}
		if prior.LastAction == "" {
			prior.LastAction = "list"
		}
		prior.PluginID = id
		prior.Version = discovered.Version
		prior.PinnedVersion = discovered.PinnedVersion
		prior.PolicyStatus = discovered.PolicyStatus
		prior.Installed = discovered.Installed
		prior.Enabled = discovered.Enabled
		prior.DiagnosticsCount = discovered.DiagnosticsCount
		state.Summaries[id] = prior
	}

	return s.Store.Save(state)
}

func (s *Service) recordAction(event PluginAction) error {
	event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	event.Action = strings.TrimSpace(strings.ToLower(event.Action))
	event.PluginID = strings.TrimSpace(event.PluginID)
	event.RequestedVersion = strings.TrimSpace(event.RequestedVersion)
	event.PreviousVersion = strings.TrimSpace(event.PreviousVersion)
	event.Version = strings.TrimSpace(event.Version)
	event.Status = strings.TrimSpace(strings.ToLower(event.Status))
	event.Detail = strings.TrimSpace(event.Detail)

	state, err := s.Store.Load()
	if err != nil {
		return err
	}
	state.History = append(state.History, event)
	return s.Store.Save(state)
}

func (s *Service) collectValidationDiagnostics(state PolicyState, runtimeErr error) map[string][]ValidationDiagnostic {
	if state.Diagnostics == nil {
		state.Diagnostics = map[string][]ValidationDiagnostic{}
	}

	diagnostics := map[string][]ValidationDiagnostic{}
	appendDiag := func(pluginID, severity, code, message string) {
		pluginID = strings.TrimSpace(pluginID)
		if pluginID == "" {
			return
		}
		diagnostics[pluginID] = append(diagnostics[pluginID], ValidationDiagnostic{
			PluginID: pluginID,
			Severity: severity,
			Code:     code,
			Message:  message,
		})
	}

	if runtimeErr != nil {
		appendDiag("_runtime", "error", "runtime_invalid", runtimeErr.Error())
	}

	entries, err := os.ReadDir(s.InstallRoot)
	if err == nil {
		idDirs := map[string][]string{}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			dirPath := filepath.Join(s.InstallRoot, entry.Name())
			manifestPath, ok := findManifestPath(dirPath)
			if !ok {
				appendDiag(entry.Name(), "error", "missing_manifest", fmt.Sprintf("No plugin manifest found in %q", dirPath))
				continue
			}
			manifest, loadErr := loadManifest(manifestPath)
			if loadErr != nil {
				appendDiag(entry.Name(), "error", "manifest_decode", fmt.Sprintf("Manifest decode failed: %v", loadErr))
				continue
			}
			normalizeManifest(&manifest)
			if validateErr := ValidateManifest(manifest); validateErr != nil {
				appendDiag(entry.Name(), "error", "manifest_invalid", validateErr.Error())
				continue
			}
			idDirs[manifest.ID] = append(idDirs[manifest.ID], entry.Name())
			if manifest.ID != entry.Name() {
				appendDiag(manifest.ID, "warning", "directory_id_mismatch", fmt.Sprintf("Manifest id %q differs from directory %q", manifest.ID, entry.Name()))
			}
		}
		for id, dirs := range idDirs {
			if len(dirs) <= 1 {
				continue
			}
			sort.Strings(dirs)
			appendDiag(id, "error", "duplicate_plugin_id", fmt.Sprintf("Duplicate plugin id across install directories: %s", strings.Join(dirs, ", ")))
		}
	}

	known, _ := s.discoveredPluginIDs()
	for _, id := range state.Installed {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, ok := known[trimmed]; !ok {
			appendDiag(trimmed, "warning", "policy_installed_missing", "Plugin is listed as installed in policy but missing from install root")
		}
	}
	for _, id := range state.Enabled {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, ok := known[trimmed]; !ok {
			appendDiag(trimmed, "warning", "policy_enabled_missing", "Plugin is listed as enabled in policy but missing from install root")
		}
	}
	for _, id := range state.Disabled {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, ok := known[trimmed]; !ok {
			appendDiag(trimmed, "warning", "policy_disabled_missing", "Plugin is listed as disabled in policy but missing from install root")
		}
	}
	for id, pinnedVersion := range state.Pins {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		if _, ok := known[trimmed]; !ok {
			appendDiag(trimmed, "warning", "policy_pin_missing", "Plugin is pinned in policy but missing from install root")
			continue
		}
		installedVersion, versionErr := installedPluginVersion(filepath.Join(s.InstallRoot, trimmed))
		if versionErr != nil {
			appendDiag(trimmed, "warning", "installed_version_unknown", fmt.Sprintf("Unable to read installed version: %v", versionErr))
			continue
		}
		if strings.TrimSpace(installedVersion) != strings.TrimSpace(pinnedVersion) {
			appendDiag(trimmed, "warning", "pinned_version_mismatch", fmt.Sprintf("Pinned version %q differs from installed version %q", strings.TrimSpace(pinnedVersion), strings.TrimSpace(installedVersion)))
		}
	}

	for id := range diagnostics {
		sort.Slice(diagnostics[id], func(i, j int) bool {
			if diagnostics[id][i].Severity != diagnostics[id][j].Severity {
				return diagnostics[id][i].Severity < diagnostics[id][j].Severity
			}
			if diagnostics[id][i].Code != diagnostics[id][j].Code {
				return diagnostics[id][i].Code < diagnostics[id][j].Code
			}
			return diagnostics[id][i].Message < diagnostics[id][j].Message
		})
	}

	return diagnostics
}

func mergePluginStateSummaries(
	existing map[string]PluginStateSummary,
	plugs []InstalledPlugin,
	diagnostics map[string][]ValidationDiagnostic,
) map[string]PluginStateSummary {
	out := map[string]PluginStateSummary{}
	for id, summary := range existing {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		summary.PluginID = trimmed
		summary.Version = strings.TrimSpace(summary.Version)
		summary.PinnedVersion = strings.TrimSpace(summary.PinnedVersion)
		summary.PolicyStatus = strings.TrimSpace(summary.PolicyStatus)
		summary.Installed = false
		summary.Enabled = false
		summary.DiagnosticsCount = len(diagnostics[trimmed])
		out[trimmed] = summary
	}

	for _, plugin := range plugs {
		id := strings.TrimSpace(plugin.Manifest.ID)
		if id == "" {
			continue
		}
		summary := out[id]
		summary.PluginID = id
		summary.Version = strings.TrimSpace(plugin.Manifest.Version)
		summary.Installed = true
		summary.Enabled = plugin.Enabled
		summary.DiagnosticsCount = len(diagnostics[id])
		out[id] = summary
	}

	for id, diags := range diagnostics {
		summary := out[id]
		summary.PluginID = id
		summary.DiagnosticsCount = len(diags)
		out[id] = summary
	}

	return out
}

func applyPolicyFieldsToSummaries(
	summaries map[string]PluginStateSummary,
	state PolicyState,
) map[string]PluginStateSummary {
	if summaries == nil {
		summaries = map[string]PluginStateSummary{}
	}

	enabledSet := make(map[string]struct{}, len(state.Enabled))
	disabledSet := make(map[string]struct{}, len(state.Disabled))
	for _, id := range state.Enabled {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			enabledSet[trimmed] = struct{}{}
		}
	}
	for _, id := range state.Disabled {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			disabledSet[trimmed] = struct{}{}
		}
	}

	for id, version := range state.Pins {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		summary := summaries[trimmed]
		summary.PluginID = trimmed
		summary.PinnedVersion = strings.TrimSpace(version)
		summaries[trimmed] = summary
	}

	for id, summary := range summaries {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		summary.PluginID = trimmed
		if _, ok := disabledSet[trimmed]; ok {
			summary.PolicyStatus = "disabled"
		} else if _, ok := enabledSet[trimmed]; ok {
			summary.PolicyStatus = "enabled"
		} else {
			summary.PolicyStatus = "default"
		}
		summaries[trimmed] = summary
	}

	return summaries
}

func buildPluginPolicySummary(
	state PolicyState,
	plugs []InstalledPlugin,
	diagnostics map[string][]ValidationDiagnostic,
) map[string]PluginPolicySummary {
	installedSet := make(map[string]struct{}, len(state.Installed))
	enabledSet := make(map[string]struct{}, len(state.Enabled))
	disabledSet := make(map[string]struct{}, len(state.Disabled))

	for _, id := range state.Installed {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			installedSet[trimmed] = struct{}{}
		}
	}
	for _, id := range state.Enabled {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			enabledSet[trimmed] = struct{}{}
		}
	}
	for _, id := range state.Disabled {
		trimmed := strings.TrimSpace(id)
		if trimmed != "" {
			disabledSet[trimmed] = struct{}{}
		}
	}

	out := map[string]PluginPolicySummary{}
	for _, plugin := range plugs {
		id := strings.TrimSpace(plugin.Manifest.ID)
		if id == "" {
			continue
		}
		_, policyInstalled := installedSet[id]
		_, policyEnabled := enabledSet[id]
		_, policyDisabled := disabledSet[id]
		diagCount := len(diagnostics[id])
		out[id] = PluginPolicySummary{
			PluginID:         id,
			PolicyInstalled:  policyInstalled,
			PolicyEnabled:    policyEnabled,
			PolicyDisabled:   policyDisabled,
			PinnedVersion:    strings.TrimSpace(state.Pins[id]),
			InstalledVersion: strings.TrimSpace(plugin.Manifest.Version),
			EffectiveEnabled: plugin.Enabled,
			HasDiagnostics:   diagCount > 0,
			DiagnosticsCount: diagCount,
		}
	}

	for id, version := range state.Pins {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			continue
		}
		summary := out[trimmed]
		summary.PluginID = trimmed
		summary.PinnedVersion = strings.TrimSpace(version)
		summary.DiagnosticsCount = len(diagnostics[trimmed])
		summary.HasDiagnostics = summary.DiagnosticsCount > 0
		if _, ok := installedSet[trimmed]; ok {
			summary.PolicyInstalled = true
		}
		if _, ok := enabledSet[trimmed]; ok {
			summary.PolicyEnabled = true
		}
		if _, ok := disabledSet[trimmed]; ok {
			summary.PolicyDisabled = true
		}
		out[trimmed] = summary
	}

	return out
}

func (s *Service) discoveredPluginIDs() (map[string]struct{}, error) {
	runtime, err := LoadRuntime(s.InstallRoot, Policy{})
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]struct{}{}, nil
		}
		return nil, err
	}
	known := make(map[string]struct{}, len(runtime.plugins))
	for _, p := range runtime.plugins {
		known[p.Manifest.ID] = struct{}{}
	}
	return known, nil
}

func removeID(list []string, target string) []string {
	target = strings.TrimSpace(target)
	if target == "" {
		return list
	}
	out := make([]string, 0, len(list))
	for _, id := range list {
		if strings.TrimSpace(id) == target {
			continue
		}
		out = append(out, id)
	}
	return out
}

func splitPluginRef(pluginRef string) (string, string) {
	trimmed := strings.TrimSpace(pluginRef)
	if trimmed == "" {
		return "", ""
	}
	parts := strings.SplitN(trimmed, "@", 2)
	id := strings.TrimSpace(parts[0])
	if len(parts) == 1 {
		return id, ""
	}
	return id, strings.TrimSpace(parts[1])
}

func installedPluginVersion(pluginDir string) (string, error) {
	manifestPath, ok := findManifestPath(pluginDir)
	if !ok {
		return "", fmt.Errorf("plugin manifest not found in %q", pluginDir)
	}
	manifest, err := loadManifest(manifestPath)
	if err != nil {
		return "", fmt.Errorf("load manifest %q: %w", manifestPath, err)
	}
	normalizeManifest(&manifest)
	if err := ValidateManifest(manifest); err != nil {
		return "", fmt.Errorf("invalid manifest %q: %w", manifestPath, err)
	}
	return manifest.Version, nil
}

func compareVersions(left, right string) int {
	ln := normalizeVersion(left)
	rn := normalizeVersion(right)
	lparts := strings.Split(ln, ".")
	rparts := strings.Split(rn, ".")
	maxParts := len(lparts)
	if len(rparts) > maxParts {
		maxParts = len(rparts)
	}
	for i := 0; i < maxParts; i++ {
		lpart := "0"
		rpart := "0"
		if i < len(lparts) {
			lpart = lparts[i]
		}
		if i < len(rparts) {
			rpart = rparts[i]
		}
		lnum, ldigits, lok := parseNumericPrefix(lpart)
		rnum, rdigits, rok := parseNumericPrefix(rpart)
		if lok && rok {
			if lnum != rnum {
				if lnum > rnum {
					return 1
				}
				return -1
			}
			lrest := lpart[ldigits:]
			rrest := rpart[rdigits:]
			if lrest != rrest {
				if lrest > rrest {
					return 1
				}
				return -1
			}
			continue
		}
		if lpart != rpart {
			if lpart > rpart {
				return 1
			}
			return -1
		}
	}
	return 0
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	if v == "" {
		return "0"
	}
	return v
}

func parseNumericPrefix(value string) (int, int, bool) {
	if value == "" {
		return 0, 0, true
	}
	i := 0
	for i < len(value) && value[i] >= '0' && value[i] <= '9' {
		i++
	}
	if i == 0 {
		return 0, 0, false
	}
	n, err := strconv.Atoi(value[:i])
	if err != nil {
		return 0, 0, false
	}
	return n, i, true
}

func copyDir(src, dst string) error {
	st, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !st.IsDir() {
		return fmt.Errorf("source %q is not a directory", src)
	}

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}

	return filepath.WalkDir(src, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink entries are not supported in plugin install source: %q", path)
		}

		return copyFile(path, target, info.Mode().Perm())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}
