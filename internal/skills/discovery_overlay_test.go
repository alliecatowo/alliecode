package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alliecatowo/alliecode/internal/plugins"
)

func TestManagerDiscoverSkillDirsForPathsAndAdd(t *testing.T) {
	workspace := t.TempDir()
	nested := filepath.Join(workspace, "apps", "api")
	skillsDir := filepath.Join(nested, ".alliecode", "skills")
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("mkdir skills dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(skillsDir, "nested.md"), []byte("---\nname: nested\n---\nnested prompt"), 0o644); err != nil {
		t.Fatalf("write nested skill: %v", err)
	}

	mgr := NewManager([]string{filepath.Join(workspace, ".alliecode", "skills")})
	found, err := mgr.DiscoverSkillDirsForPaths([]string{filepath.Join(nested, "main.go")}, workspace)
	if err != nil {
		t.Fatalf("DiscoverSkillDirsForPaths() error = %v", err)
	}
	if len(found) != 1 || found[0] != skillsDir {
		t.Fatalf("unexpected discovered dirs: %#v", found)
	}

	if err := mgr.AddDiscoveredDirectories(found); err != nil {
		t.Fatalf("AddDiscoveredDirectories() error = %v", err)
	}
	if _, ok := mgr.Get("nested"); !ok {
		t.Fatalf("expected dynamically loaded nested skill")
	}

	again, err := mgr.DiscoverSkillDirsForPaths([]string{filepath.Join(nested, "other.go")}, workspace)
	if err != nil {
		t.Fatalf("DiscoverSkillDirsForPaths(second) error = %v", err)
	}
	if len(again) != 0 {
		t.Fatalf("expected previously discovered dirs to be skipped, got %#v", again)
	}
}

func TestMergeSkillsWithPluginCommandsDetailedMetadataAndOverlay(t *testing.T) {
	base := []*Skill{
		{Name: "shared", Description: "file shared", FilePath: "/skills/shared.md", Enabled: true, Source: "file"},
		{Name: "local", Description: "file local", FilePath: "/skills/local.md", Enabled: true, Source: "file"},
	}

	pluginCommands := []plugins.ResolvedCommand{
		{PluginID: "alpha", Name: "shared", Description: "plugin shared", Usage: "/shared"},
		{PluginID: "beta", Name: "shared", Description: "plugin shared beta", Usage: "/shared"},
		{PluginID: "alpha", Name: "plugin-only", Description: "plugin only", Usage: "/plugin-only"},
	}

	detailed := MergeSkillsWithPluginCommandsDetailed(base, pluginCommands, PreferSkillFiles)
	meta, ok := detailed.Metadata["shared"]
	if !ok || meta.Source != "file" || meta.WinnerSource != "/skills/shared.md" {
		t.Fatalf("unexpected shared metadata: %#v", meta)
	}
	overlay := detailed.PluginOverlay["shared"]
	if len(overlay) != 2 || overlay[0] != "alpha" || overlay[1] != "beta" {
		t.Fatalf("unexpected shared overlay list: %#v", overlay)
	}

	pluginMeta, ok := detailed.Metadata["plugin-only"]
	if !ok || pluginMeta.Source != "plugin" || pluginMeta.PluginID != "alpha" {
		t.Fatalf("unexpected plugin-only metadata: %#v", pluginMeta)
	}
}
