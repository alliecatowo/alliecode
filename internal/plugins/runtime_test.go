package plugins

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRuntimeDiscovery(t *testing.T) {
	root := t.TempDir()

	writePluginManifest(t, root, "alpha", `id: alpha
name: Alpha
version: 1.0.0
commands:
  - name: sync
    description: sync workspace
    usage: /sync
tools:
  - name: alpha.search
    description: search in alpha
`)

	writePluginManifest(t, root, "beta", `id: beta
name: Beta
version: 2.1.0
commands:
  - name: analyze
tools:
  - name: beta.fetch
`)

	rt, err := LoadRuntime(root, Policy{})
	if err != nil {
		t.Fatalf("LoadRuntime() error = %v", err)
	}

	plugins := rt.Plugins()
	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}
	if plugins[0].Manifest.ID != "alpha" || plugins[1].Manifest.ID != "beta" {
		t.Fatalf("unexpected plugin order: %q, %q", plugins[0].Manifest.ID, plugins[1].Manifest.ID)
	}
	if !plugins[0].Enabled || !plugins[1].Enabled {
		t.Fatalf("expected all plugins enabled by default")
	}

	commands := rt.Commands()
	if len(commands) != 2 {
		t.Fatalf("expected 2 resolved commands, got %d", len(commands))
	}
	if commands[0].PluginID != "beta" || commands[0].Name != "analyze" {
		t.Fatalf("unexpected first command: %+v", commands[0])
	}
	if commands[1].PluginID != "alpha" || commands[1].Name != "sync" {
		t.Fatalf("unexpected second command: %+v", commands[1])
	}

	tools := rt.Tools()
	if len(tools) != 2 {
		t.Fatalf("expected 2 resolved tools, got %d", len(tools))
	}
	if tools[0].PluginID != "alpha" || tools[0].Name != "alpha.search" {
		t.Fatalf("unexpected first tool: %+v", tools[0])
	}
	if tools[1].PluginID != "beta" || tools[1].Name != "beta.fetch" {
		t.Fatalf("unexpected second tool: %+v", tools[1])
	}
}

func TestLoadRuntimeMalformedManifest(t *testing.T) {
	root := t.TempDir()

	writePluginManifest(t, root, "good", `id: good
version: 1.0.0
`)
	writePluginManifest(t, root, "missing-id", `version: 1.0.0
`)
	writePluginManifest(t, root, "unknown-field", `id: bad
version: 1.0.0
surprise: true
`)

	_, err := LoadRuntime(root, Policy{})
	if err == nil {
		t.Fatalf("expected malformed manifest error")
	}
}

func TestLoadRuntimeManifestNormalizationAndDeterministicListings(t *testing.T) {
	root := t.TempDir()

	writePluginManifest(t, root, "beta", `id: beta
name: "  Beta  "
version: " 1.0.0 "
commands:
  - name: " zeta "
  - name: " alpha "
tools:
  - name: " z-tool "
  - name: " a-tool "
`)
	writePluginManifest(t, root, "alpha", `id: alpha
name: Alpha
version: 1.0.0
commands:
  - name: common
tools:
  - name: common.tool
`)

	rt, err := LoadRuntime(root, Policy{})
	if err != nil {
		t.Fatalf("LoadRuntime() error = %v", err)
	}

	plugins := rt.Plugins()
	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}
	if plugins[1].Manifest.Name != "Beta" || plugins[1].Manifest.Version != "1.0.0" {
		t.Fatalf("expected normalized manifest fields, got %+v", plugins[1].Manifest)
	}

	commands := rt.Commands()
	if len(commands) != 3 {
		t.Fatalf("expected 3 commands, got %+v", commands)
	}
	if commands[0].Name != "alpha" || commands[1].Name != "common" || commands[2].Name != "zeta" {
		t.Fatalf("expected deterministic sorted commands, got %+v", commands)
	}

	tools := rt.Tools()
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %+v", tools)
	}
	if tools[0].Name != "a-tool" || tools[1].Name != "common.tool" || tools[2].Name != "z-tool" {
		t.Fatalf("expected deterministic sorted tools, got %+v", tools)
	}
}

func TestLoadRuntimePolicyFiltering(t *testing.T) {
	root := t.TempDir()

	writePluginManifest(t, root, "alpha", `id: alpha
version: 1.0.0
commands:
  - name: alpha-cmd
tools:
  - name: alpha-tool
`)
	writePluginManifest(t, root, "beta", `id: beta
version: 1.0.0
commands:
  - name: beta-cmd
tools:
  - name: beta-tool
`)

	rt, err := LoadRuntime(root, Policy{Enabled: []string{"beta"}})
	if err != nil {
		t.Fatalf("LoadRuntime enabled filter error = %v", err)
	}
	plugins := rt.Plugins()
	if len(plugins) != 2 {
		t.Fatalf("expected 2 discovered plugins, got %d", len(plugins))
	}
	if plugins[0].Enabled {
		t.Fatalf("expected alpha to be disabled by enabled allow-list")
	}
	if !plugins[1].Enabled {
		t.Fatalf("expected beta to be enabled by enabled allow-list")
	}
	if len(rt.Commands()) != 1 || rt.Commands()[0].PluginID != "beta" {
		t.Fatalf("expected only beta command, got %+v", rt.Commands())
	}
	if len(rt.Tools()) != 1 || rt.Tools()[0].PluginID != "beta" {
		t.Fatalf("expected only beta tool, got %+v", rt.Tools())
	}

	rt, err = LoadRuntime(root, Policy{Disabled: []string{"beta"}})
	if err != nil {
		t.Fatalf("LoadRuntime disabled filter error = %v", err)
	}
	if len(rt.Commands()) != 1 || rt.Commands()[0].PluginID != "alpha" {
		t.Fatalf("expected only alpha command after disabling beta, got %+v", rt.Commands())
	}
	if len(rt.Tools()) != 1 || rt.Tools()[0].PluginID != "alpha" {
		t.Fatalf("expected only alpha tool after disabling beta, got %+v", rt.Tools())
	}

	rt, err = LoadRuntime(root, Policy{Enabled: []string{" beta "}})
	if err != nil {
		t.Fatalf("LoadRuntime whitespace policy error = %v", err)
	}
	if len(rt.Commands()) != 1 || rt.Commands()[0].PluginID != "beta" {
		t.Fatalf("expected trimmed enabled policy to allow beta, got %+v", rt.Commands())
	}

	_, err = LoadRuntime(root, Policy{Enabled: []string{"missing"}})
	if err == nil {
		t.Fatalf("expected unknown plugin policy error")
	}
}

func TestValidateManifestRejectsTrimmedDuplicateNames(t *testing.T) {
	err := ValidateManifest(Manifest{
		ID:      "dup-test",
		Version: "1.0.0",
		Commands: []CommandManifest{
			{Name: "sync"},
			{Name: " sync "},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate command name") {
		t.Fatalf("expected duplicate command name error, got %v", err)
	}

	err = ValidateManifest(Manifest{
		ID:      "dup-test-tools",
		Version: "1.0.0",
		Tools: []ToolManifest{
			{Name: "search"},
			{Name: " search "},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate tool name") {
		t.Fatalf("expected duplicate tool name error, got %v", err)
	}
}

func writePluginManifest(t *testing.T, root, dirName, body string) {
	t.Helper()
	dir := filepath.Join(root, dirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	path := filepath.Join(dir, "plugin.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
