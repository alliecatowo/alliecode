package plugins

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestPolicyStateNormalizePins(t *testing.T) {
	path := filepath.Join(t.TempDir(), "policy.yaml")
	store := NewPolicyStore(path)

	state := PolicyState{
		Installed: []string{"alpha"},
		Pins: map[string]string{
			" alpha ": " 1.2.3 ",
			"":        "2.0.0",
			"beta":    "",
		},
	}
	if err := store.Save(state); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(loaded.Pins) != 1 || loaded.Pins["alpha"] != "1.2.3" {
		t.Fatalf("unexpected normalized pins: %#v", loaded.Pins)
	}
}

func TestServiceUpdateUsesPinnedVersionByDefault(t *testing.T) {
	root := t.TempDir()
	registryRoot := filepath.Join(root, "registry")
	installRoot := filepath.Join(root, "installed")
	if err := os.MkdirAll(registryRoot, 0o755); err != nil {
		t.Fatalf("mkdir registry root: %v", err)
	}

	writePluginManifest(t, registryRoot, "alpha-v1", "id: alpha\nversion: 1.0.0\n")
	writePluginManifest(t, registryRoot, "alpha-v2", "id: alpha\nversion: 2.0.0\n")

	indexPath := filepath.Join(root, "index.yaml")
	indexContent := "plugins:\n  - id: alpha\n    version: 1.0.0\n    source_path: registry/alpha-v1\n  - id: alpha\n    version: 2.0.0\n    source_path: registry/alpha-v2\n"
	if err := os.WriteFile(indexPath, []byte(indexContent), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	service := NewService(installRoot, filepath.Join(root, "policy.yaml"), indexPath)
	if err := service.Install("alpha@1.0.0"); err != nil {
		t.Fatalf("Install(alpha@1.0.0) error = %v", err)
	}

	if err := service.Update("alpha"); err != nil {
		t.Fatalf("Update(alpha) error = %v", err)
	}

	result, err := service.ListWithDiagnostics()
	if err != nil {
		t.Fatalf("ListWithDiagnostics() error = %v", err)
	}
	if got := result.Summaries["alpha"].Version; got != "1.0.0" {
		t.Fatalf("expected pinned version to remain 1.0.0, got %q", got)
	}
}

func TestServicePolicySummaryIncludesPinsAndEffectiveState(t *testing.T) {
	root := t.TempDir()
	registryRoot := filepath.Join(root, "registry")
	installRoot := filepath.Join(root, "installed")
	if err := os.MkdirAll(registryRoot, 0o755); err != nil {
		t.Fatalf("mkdir registry root: %v", err)
	}

	writePluginManifest(t, registryRoot, "alpha", "id: alpha\nversion: 1.0.0\n")
	indexPath := filepath.Join(root, "index.yaml")
	indexContent := "plugins:\n  - id: alpha\n    version: 1.0.0\n    source_path: registry/alpha\n"
	if err := os.WriteFile(indexPath, []byte(indexContent), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	service := NewService(installRoot, filepath.Join(root, "policy.yaml"), indexPath)
	if err := service.Install("alpha"); err != nil {
		t.Fatalf("Install(alpha) error = %v", err)
	}
	if err := service.PinVersion("alpha@1.0.0"); err != nil {
		t.Fatalf("PinVersion(alpha@1.0.0) error = %v", err)
	}
	if err := service.Disable("alpha"); err != nil {
		t.Fatalf("Disable(alpha) error = %v", err)
	}

	result, err := service.ListWithDiagnostics()
	if err != nil {
		t.Fatalf("ListWithDiagnostics() error = %v", err)
	}

	summary, ok := result.PolicySummary["alpha"]
	if !ok {
		t.Fatalf("expected alpha policy summary, got %#v", result.PolicySummary)
	}
	if summary.PinnedVersion != "1.0.0" || summary.EffectiveEnabled {
		t.Fatalf("unexpected policy summary: %#v", summary)
	}

	payload, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if string(payload) == "{}" {
		t.Fatalf("expected populated summary payload")
	}
}
