package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alliecatowo/alliecode/internal/plugins"
	"github.com/alliecatowo/alliecode/internal/skills"
)

func TestPluginSkillsParity_MergeDeterministicOrigins(t *testing.T) {
	baseSkills := []*skills.Skill{
		{Name: "review", Description: "from file", FilePath: "/skills/review.md", Enabled: true, Source: "file"},
	}
	pluginCommands := []plugins.ResolvedCommand{
		{PluginID: "qa", Name: "review", Description: "from plugin", Usage: "/review"},
		{PluginID: "qa", Name: "lint", Description: "lint workspace", Usage: "/lint"},
	}

	merged := skills.MergeSkillsWithPluginCommandsDetailed(baseSkills, pluginCommands, skills.PreferSkillFiles)
	if len(merged.Skills) != 2 {
		t.Fatalf("expected 2 merged skills, got %d", len(merged.Skills))
	}
	if got := merged.PluginOrigins["lint"]; got != "qa" {
		t.Fatalf("plugin origin mismatch for lint: got %q", got)
	}
	if got := merged.Metadata["review"].WinnerSource; got != "/skills/review.md" {
		t.Fatalf("winner source mismatch for review: got %q", got)
	}
}

func TestPluginSkillsParity_PolicySummaryShape(t *testing.T) {
	root := t.TempDir()
	registryRoot := filepath.Join(root, "registry")
	installRoot := filepath.Join(root, "installed")
	if err := os.MkdirAll(registryRoot, 0o755); err != nil {
		t.Fatalf("mkdir registry root: %v", err)
	}

	alphaDir := filepath.Join(registryRoot, "alpha")
	if err := os.MkdirAll(alphaDir, 0o755); err != nil {
		t.Fatalf("mkdir alpha dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(alphaDir, "plugin.yaml"), []byte("id: alpha\nversion: 1.0.0\n"), 0o644); err != nil {
		t.Fatalf("write alpha manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "index.yaml"), []byte("plugins:\n  - id: alpha\n    version: 1.0.0\n    source_path: registry/alpha\n"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	service := plugins.NewService(installRoot, filepath.Join(root, "policy.yaml"), filepath.Join(root, "index.yaml"))
	if err := service.Install("alpha"); err != nil {
		t.Fatalf("Install(alpha) error = %v", err)
	}
	if err := service.PinVersion("alpha@1.0.0"); err != nil {
		t.Fatalf("PinVersion(alpha) error = %v", err)
	}

	result, err := service.ListWithDiagnostics()
	if err != nil {
		t.Fatalf("ListWithDiagnostics() error = %v", err)
	}
	policy := result.PolicySummary["alpha"]
	if !policy.PolicyInstalled || !policy.PolicyEnabled || policy.PinnedVersion != "1.0.0" {
		t.Fatalf("unexpected policy summary: %#v", policy)
	}
}
