package plugins

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPolicyStoreSaveLoadAndRuntimePolicyFiltering(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy.yaml")
	store := NewPolicyStore(path)

	state := PolicyState{
		Installed: []string{" beta ", "alpha", "alpha"},
		Enabled:   []string{"beta", "alpha", "beta"},
		Disabled:  []string{"beta", "missing"},
	}
	if err := store.Save(state); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded.Installed) != 2 || loaded.Installed[0] != "alpha" || loaded.Installed[1] != "beta" {
		t.Fatalf("unexpected installed state: %#v", loaded.Installed)
	}
	if len(loaded.Enabled) != 1 || loaded.Enabled[0] != "alpha" {
		t.Fatalf("expected disabled entries removed from enabled, got %#v", loaded.Enabled)
	}

	runtime := loaded.RuntimePolicy([]string{"alpha"})
	if len(runtime.Enabled) != 1 || runtime.Enabled[0] != "alpha" {
		t.Fatalf("unexpected runtime enabled set: %#v", runtime.Enabled)
	}
	if len(runtime.Disabled) != 0 {
		t.Fatalf("unexpected runtime disabled set: %#v", runtime.Disabled)
	}
}

func TestServiceInstallEnableDisableList(t *testing.T) {
	root := t.TempDir()
	registryRoot := filepath.Join(root, "registry")
	installRoot := filepath.Join(root, "installed")
	if err := os.MkdirAll(registryRoot, 0o755); err != nil {
		t.Fatalf("mkdir registry root: %v", err)
	}

	writePluginManifest(t, registryRoot, "alpha", `id: alpha
version: 1.0.0
commands:
  - name: alpha-cmd
tools:
  - name: alpha-tool
`)
	writePluginManifest(t, registryRoot, "beta", `id: beta
version: 1.1.0
commands:
  - name: beta-cmd
tools:
  - name: beta-tool
`)

	indexPath := filepath.Join(root, "index.yaml")
	indexContent := `plugins:
  - id: alpha
    version: 1.0.0
    source_path: registry/alpha
  - id: beta
    version: 1.1.0
    source_path: registry/beta
`
	if err := os.WriteFile(indexPath, []byte(indexContent), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	service := NewService(installRoot, filepath.Join(root, "policy.yaml"), indexPath)

	if err := service.Install("alpha"); err != nil {
		t.Fatalf("Install(alpha) error = %v", err)
	}
	if err := service.Install("alpha"); err == nil {
		t.Fatalf("expected duplicate install error")
	}

	plugins, err := service.List()
	if err != nil {
		t.Fatalf("List() after install error = %v", err)
	}
	if len(plugins) != 1 || plugins[0].Manifest.ID != "alpha" || !plugins[0].Enabled {
		t.Fatalf("unexpected plugins after alpha install: %#v", plugins)
	}

	if err := service.Disable("alpha"); err != nil {
		t.Fatalf("Disable(alpha) error = %v", err)
	}
	plugins, err = service.List()
	if err != nil {
		t.Fatalf("List() after disable error = %v", err)
	}
	if len(plugins) != 1 || plugins[0].Enabled {
		t.Fatalf("expected alpha disabled, got %#v", plugins)
	}

	if err := service.Enable("alpha"); err != nil {
		t.Fatalf("Enable(alpha) error = %v", err)
	}

	if err := service.Install("beta"); err != nil {
		t.Fatalf("Install(beta) error = %v", err)
	}
	if err := service.Disable("beta"); err != nil {
		t.Fatalf("Disable(beta) error = %v", err)
	}

	plugins, err = service.List()
	if err != nil {
		t.Fatalf("List() final error = %v", err)
	}
	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %#v", plugins)
	}
	if plugins[0].Manifest.ID != "alpha" || !plugins[0].Enabled {
		t.Fatalf("unexpected alpha state: %#v", plugins[0])
	}
	if plugins[1].Manifest.ID != "beta" || plugins[1].Enabled {
		t.Fatalf("unexpected beta state: %#v", plugins[1])
	}
}

func TestServiceInstallUpdateRemoveWithVersionPinningAndHistory(t *testing.T) {
	root := t.TempDir()
	registryRoot := filepath.Join(root, "registry")
	installRoot := filepath.Join(root, "installed")
	if err := os.MkdirAll(registryRoot, 0o755); err != nil {
		t.Fatalf("mkdir registry root: %v", err)
	}

	writePluginManifest(t, registryRoot, "alpha-v1", `id: alpha
version: 1.0.0
commands:
  - name: alpha-cmd
`)
	writePluginManifest(t, registryRoot, "alpha-v2", `id: alpha
version: 2.0.0
commands:
  - name: alpha-cmd
`)

	indexPath := filepath.Join(root, "index.yaml")
	indexContent := `plugins:
  - id: alpha
    version: 1.0.0
    source_path: registry/alpha-v1
  - id: alpha
    version: 2.0.0
    source_path: registry/alpha-v2
`
	if err := os.WriteFile(indexPath, []byte(indexContent), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	service := NewService(installRoot, filepath.Join(root, "policy.yaml"), indexPath)
	if err := service.Install("alpha@1.0.0"); err != nil {
		t.Fatalf("Install(alpha@1.0.0) error = %v", err)
	}

	result, err := service.ListWithDiagnostics()
	if err != nil {
		t.Fatalf("ListWithDiagnostics() after install error = %v", err)
	}
	if got := result.Summaries["alpha"].Version; got != "1.0.0" {
		t.Fatalf("expected installed version 1.0.0, got %q", got)
	}

	if err := service.Update("alpha"); err != nil {
		t.Fatalf("Update(alpha) error = %v", err)
	}
	result, err = service.ListWithDiagnostics()
	if err != nil {
		t.Fatalf("ListWithDiagnostics() after update error = %v", err)
	}
	if got := result.Summaries["alpha"].Version; got != "1.0.0" {
		t.Fatalf("expected pinned version 1.0.0 to remain installed, got %q", got)
	}

	if err := service.UnpinVersion("alpha"); err != nil {
		t.Fatalf("UnpinVersion(alpha) error = %v", err)
	}
	if err := service.Update("alpha"); err != nil {
		t.Fatalf("Update(alpha) after unpin error = %v", err)
	}
	result, err = service.ListWithDiagnostics()
	if err != nil {
		t.Fatalf("ListWithDiagnostics() after unpinned update error = %v", err)
	}
	if got := result.Summaries["alpha"].Version; got != "2.0.0" {
		t.Fatalf("expected updated version 2.0.0 after unpin, got %q", got)
	}

	if err := service.Remove("alpha"); err != nil {
		t.Fatalf("Remove(alpha) error = %v", err)
	}
	result, err = service.ListWithDiagnostics()
	if err != nil {
		t.Fatalf("ListWithDiagnostics() after remove error = %v", err)
	}
	if len(result.Plugins) != 0 {
		t.Fatalf("expected no installed plugins after remove, got %#v", result.Plugins)
	}
	if summary := result.Summaries["alpha"]; summary.Installed {
		t.Fatalf("expected removed plugin summary to be installed=false, got %#v", summary)
	}

	if len(result.History) < 3 {
		t.Fatalf("expected at least 3 history events, got %#v", result.History)
	}
	last := result.History[len(result.History)-1]
	if last.Action != "remove" || last.PluginID != "alpha" || last.Status == "" {
		t.Fatalf("unexpected last history event: %#v", last)
	}
}

func TestServiceListWithDiagnosticsAndPersistedSummaries(t *testing.T) {
	root := t.TempDir()
	registryRoot := filepath.Join(root, "registry")
	installRoot := filepath.Join(root, "installed")
	if err := os.MkdirAll(registryRoot, 0o755); err != nil {
		t.Fatalf("mkdir registry root: %v", err)
	}

	writePluginManifest(t, registryRoot, "alpha", `id: alpha
version: 1.0.0
commands:
  - name: alpha-cmd
`)

	indexPath := filepath.Join(root, "index.yaml")
	indexContent := `plugins:
  - id: alpha
    version: 1.0.0
    source_path: registry/alpha
`
	if err := os.WriteFile(indexPath, []byte(indexContent), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	service := NewService(installRoot, filepath.Join(root, "policy.yaml"), indexPath)
	if err := service.Install("alpha"); err != nil {
		t.Fatalf("Install(alpha) error = %v", err)
	}

	result, err := service.ListWithDiagnostics()
	if err != nil {
		t.Fatalf("ListWithDiagnostics() error = %v", err)
	}
	if len(result.Plugins) != 1 || result.Plugins[0].Manifest.ID != "alpha" {
		t.Fatalf("unexpected plugin list: %#v", result.Plugins)
	}
	summary, ok := result.Summaries["alpha"]
	if !ok {
		t.Fatalf("expected alpha summary, got %#v", result.Summaries)
	}
	if !summary.Installed || !summary.Enabled || summary.Version != "1.0.0" {
		t.Fatalf("unexpected alpha summary: %#v", summary)
	}
	if summary.LastAction == "" || summary.LastUpdatedAt == "" {
		t.Fatalf("expected persisted action metadata, got %#v", summary)
	}

	brokenDir := filepath.Join(installRoot, "broken")
	if err := os.MkdirAll(brokenDir, 0o755); err != nil {
		t.Fatalf("mkdir broken plugin dir: %v", err)
	}

	result, err = service.ListWithDiagnostics()
	if err != nil {
		t.Fatalf("ListWithDiagnostics() with broken dir error = %v", err)
	}
	brokenDiagnostics, ok := result.Diagnostics["broken"]
	if !ok || len(brokenDiagnostics) == 0 {
		t.Fatalf("expected diagnostics for broken plugin directory, got %#v", result.Diagnostics)
	}
	if brokenDiagnostics[0].Code != "missing_manifest" {
		t.Fatalf("expected missing_manifest diagnostic, got %#v", brokenDiagnostics[0])
	}
}

func TestServiceListWithDiagnosticsIncludesRuntimeErrors(t *testing.T) {
	root := t.TempDir()
	installRoot := filepath.Join(root, "installed")
	if err := os.MkdirAll(filepath.Join(installRoot, "alpha"), 0o755); err != nil {
		t.Fatalf("mkdir alpha: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(installRoot, "alpha-dup"), 0o755); err != nil {
		t.Fatalf("mkdir alpha-dup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(installRoot, "alpha", "plugin.yaml"), []byte("id: alpha\nversion: 1.0.0\n"), 0o644); err != nil {
		t.Fatalf("write alpha manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(installRoot, "alpha-dup", "plugin.yaml"), []byte("id: alpha\nversion: 1.1.0\n"), 0o644); err != nil {
		t.Fatalf("write alpha-dup manifest: %v", err)
	}

	service := NewService(installRoot, filepath.Join(root, "policy.yaml"), filepath.Join(root, "index.yaml"))
	result, err := service.ListWithDiagnostics()
	if err != nil {
		t.Fatalf("ListWithDiagnostics() error = %v", err)
	}
	diags, ok := result.Diagnostics["_runtime"]
	if !ok || len(diags) == 0 {
		t.Fatalf("expected runtime diagnostics, got %#v", result.Diagnostics)
	}
	if diags[0].Code != "runtime_invalid" || !strings.Contains(diags[0].Message, "duplicated") {
		t.Fatalf("unexpected runtime diagnostic: %#v", diags[0])
	}
}

func TestMarketplaceResolveVersionSelection(t *testing.T) {
	root := t.TempDir()
	indexPath := filepath.Join(root, "index.yaml")
	indexContent := `plugins:
  - id: alpha
    version: 1.0.0
    source_path: ./alpha-v1
  - id: alpha
    version: 1.2.0
    source_path: ./alpha-v12
  - id: beta
    version: 0.9.0
    source_path: ./beta
`
	if err := os.WriteFile(indexPath, []byte(indexContent), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	m := &Marketplace{IndexPath: indexPath}
	entry, err := m.Resolve("alpha")
	if err != nil {
		t.Fatalf("Resolve(alpha) error = %v", err)
	}
	if entry.Version != "1.2.0" {
		t.Fatalf("expected latest alpha version, got %#v", entry)
	}

	pinned, err := m.Resolve("alpha@1.0.0")
	if err != nil {
		t.Fatalf("Resolve(alpha@1.0.0) error = %v", err)
	}
	if pinned.Version != "1.0.0" {
		t.Fatalf("expected pinned alpha version, got %#v", pinned)
	}

	if _, err := m.Resolve("alpha@9.9.9"); err == nil {
		t.Fatalf("expected pinned version lookup failure")
	}
}
