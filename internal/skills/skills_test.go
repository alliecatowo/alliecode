package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alliecatowo/alliecode/internal/plugins"
)

func TestParseSkillFileFrontmatterMetadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "commit.md")
	content := "---\nname: commit\ndescription: Commit helper\ntags: [git, workflow]\ntools: [Bash, Read]\nenabled: false\n---\nUse git status and draft a commit.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write skill file: %v", err)
	}

	skill, err := parseSkillFile(path)
	if err != nil {
		t.Fatalf("parse skill: %v", err)
	}

	if skill.Name != "commit" {
		t.Fatalf("name mismatch: got %q", skill.Name)
	}
	if skill.Description != "Commit helper" {
		t.Fatalf("description mismatch: got %q", skill.Description)
	}
	if len(skill.Tags) != 2 || skill.Tags[0] != "git" || skill.Tags[1] != "workflow" {
		t.Fatalf("tags mismatch: got %#v", skill.Tags)
	}
	if len(skill.Tools) != 2 || skill.Tools[0] != "Bash" || skill.Tools[1] != "Read" {
		t.Fatalf("tools mismatch: got %#v", skill.Tools)
	}
	if skill.Enabled {
		t.Fatalf("expected disabled skill")
	}
	if skill.Prompt != "Use git status and draft a commit." {
		t.Fatalf("prompt mismatch: got %q", skill.Prompt)
	}
}

func TestParseSkillFileFrontmatterAliasesAndScalarLists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ops.md")
	content := "---\ntitle: ops\nsummary: ops helper\ntag: automation\ntool: Bash\n---\nRun an operations checklist.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write skill file: %v", err)
	}

	skill, err := parseSkillFile(path)
	if err != nil {
		t.Fatalf("parse skill: %v", err)
	}

	if skill.Name != "ops" {
		t.Fatalf("name mismatch: got %q", skill.Name)
	}
	if skill.Description != "ops helper" {
		t.Fatalf("description mismatch: got %q", skill.Description)
	}
	if len(skill.Tags) != 1 || skill.Tags[0] != "automation" {
		t.Fatalf("tags mismatch: got %#v", skill.Tags)
	}
	if len(skill.Tools) != 1 || skill.Tools[0] != "Bash" {
		t.Fatalf("tools mismatch: got %#v", skill.Tools)
	}
}

func TestParseSkillFileHeadingMetadata(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "review.md")
	content := "# review\ndescription: PR reviewer\ntags: review, code-quality\nenabled: true\n\nFocus on correctness and edge cases.\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write skill file: %v", err)
	}

	skill, err := parseSkillFile(path)
	if err != nil {
		t.Fatalf("parse skill: %v", err)
	}

	if skill.Name != "review" {
		t.Fatalf("name mismatch: got %q", skill.Name)
	}
	if skill.Description != "PR reviewer" {
		t.Fatalf("description mismatch: got %q", skill.Description)
	}
	if len(skill.Tags) != 2 || skill.Tags[0] != "review" || skill.Tags[1] != "code-quality" {
		t.Fatalf("tags mismatch: got %#v", skill.Tags)
	}
	if !skill.Enabled {
		t.Fatalf("expected enabled skill")
	}
	if skill.Prompt != "Focus on correctness and edge cases." {
		t.Fatalf("prompt mismatch: got %q", skill.Prompt)
	}
}

func TestManagerDiscoveryDefaultWorkspaceAndGlobal(t *testing.T) {
	workspace := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWD)
	}()
	if err := os.Chdir(workspace); err != nil {
		t.Fatalf("chdir workspace: %v", err)
	}

	workspaceSkills := filepath.Join(workspace, ".alliecode", "skills")
	globalSkills := filepath.Join(home, ".alliecode", "skills")
	if err := os.MkdirAll(workspaceSkills, 0o755); err != nil {
		t.Fatalf("mkdir workspace skills: %v", err)
	}
	if err := os.MkdirAll(globalSkills, 0o755); err != nil {
		t.Fatalf("mkdir global skills: %v", err)
	}

	if err := os.WriteFile(filepath.Join(workspaceSkills, "shared.md"), []byte("---\nname: shared\ndescription: workspace\n---\nworkspace prompt"), 0o644); err != nil {
		t.Fatalf("write workspace skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(globalSkills, "shared.md"), []byte("---\nname: shared\ndescription: global\n---\nglobal prompt"), 0o644); err != nil {
		t.Fatalf("write global shared skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(globalSkills, "global-only.md"), []byte("---\nname: global-only\n---\nglobal only"), 0o644); err != nil {
		t.Fatalf("write global only skill: %v", err)
	}

	mgr := NewManager(nil)
	if err := mgr.Load(); err != nil {
		t.Fatalf("load skills: %v", err)
	}

	shared, ok := mgr.Get("shared")
	if !ok {
		t.Fatalf("expected shared skill")
	}
	if shared.Description != "workspace" {
		t.Fatalf("expected workspace skill precedence, got %q", shared.Description)
	}

	if _, ok := mgr.Get("global-only"); !ok {
		t.Fatalf("expected global-only skill")
	}
}

func TestManagerDiscoveryUsesDeterministicPathPrecedenceWithinDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "00-first"), 0o755); err != nil {
		t.Fatalf("mkdir first: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "99-second"), 0o755); err != nil {
		t.Fatalf("mkdir second: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "99-second", "shared.md"), []byte("---\nname: shared\ndescription: second\n---\nsecond"), 0o644); err != nil {
		t.Fatalf("write second: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "00-first", "shared.md"), []byte("---\nname: shared\ndescription: first\n---\nfirst"), 0o644); err != nil {
		t.Fatalf("write first: %v", err)
	}

	mgr := NewManager([]string{dir})
	if err := mgr.Load(); err != nil {
		t.Fatalf("load skills: %v", err)
	}

	shared, ok := mgr.Get("shared")
	if !ok {
		t.Fatalf("expected shared skill")
	}
	if shared.Description != "first" {
		t.Fatalf("expected lexicographically first path precedence, got %q", shared.Description)
	}
}

func TestManagerFilterAndDisabledExecution(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "nested", "review.md"), []byte("---\nname: reviewer\ntags: [review]\n---\nreview prompt"), 0o644); err != nil {
		t.Fatalf("write review skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "disabled.md"), []byte("---\nname: deploy\ntags: [ops]\nenabled: false\n---\ndeploy prompt"), 0o644); err != nil {
		t.Fatalf("write disabled skill: %v", err)
	}

	mgr := NewManager([]string{dir})
	if err := mgr.Load(); err != nil {
		t.Fatalf("load skills: %v", err)
	}

	filtered := mgr.Filter(FilterOptions{Tag: "review"})
	if len(filtered) != 1 || filtered[0].Name != "reviewer" {
		t.Fatalf("tag filter mismatch: got %#v", filtered)
	}

	nameFiltered := mgr.Filter(FilterOptions{NameContains: "ploy", IncludeDisabled: true})
	if len(nameFiltered) != 1 || nameFiltered[0].Name != "deploy" {
		t.Fatalf("name filter mismatch: got %#v", nameFiltered)
	}

	exactNameFiltered := mgr.Filter(FilterOptions{ExactName: "REVIEWER"})
	if len(exactNameFiltered) != 1 || exactNameFiltered[0].Name != "reviewer" {
		t.Fatalf("exact name filter mismatch: got %#v", exactNameFiltered)
	}

	toolFiltered := mgr.Filter(FilterOptions{Tool: "bash", IncludeDisabled: true})
	if len(toolFiltered) != 0 {
		t.Fatalf("tool filter should be empty before tool metadata setup: got %#v", toolFiltered)
	}

	defaultFiltered := mgr.Filter(FilterOptions{})
	if len(defaultFiltered) != 1 || defaultFiltered[0].Name != "reviewer" {
		t.Fatalf("default filter should exclude disabled: got %#v", defaultFiltered)
	}

	if _, err := mgr.Execute("deploy", ""); err == nil {
		t.Fatalf("expected execute error for disabled skill")
	}
}

func TestManagerFilterByTool(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte("---\nname: a\ntools: [Bash, Read]\n---\na"), 0o644); err != nil {
		t.Fatalf("write a skill: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.md"), []byte("---\nname: b\ntool: browser\n---\nb"), 0o644); err != nil {
		t.Fatalf("write b skill: %v", err)
	}

	mgr := NewManager([]string{dir})
	if err := mgr.Load(); err != nil {
		t.Fatalf("load skills: %v", err)
	}

	filtered := mgr.Filter(FilterOptions{Tool: "read"})
	if len(filtered) != 1 || filtered[0].Name != "a" {
		t.Fatalf("tool filter mismatch: got %#v", filtered)
	}
}

func TestMergeSkillsWithPluginCommandsPrecedence(t *testing.T) {
	base := []*Skill{
		{Name: "shared", Description: "skill-shared", FilePath: "/skills/shared.md", Enabled: true},
		{Name: "local-only", Description: "skill-local", FilePath: "/skills/local-only.md", Enabled: true},
	}

	pluginCommands := []plugins.ResolvedCommand{
		{PluginID: "alpha", Name: "shared", Description: "plugin-shared", Usage: "/shared"},
		{PluginID: "alpha", Name: "plugin-only", Description: "plugin-only", Usage: "/plugin-only"},
	}

	skillsFirst := MergeSkillsWithPluginCommands(base, pluginCommands, PreferSkillFiles)
	sharedSkill, ok := findSkillByName(skillsFirst, "shared")
	if !ok {
		t.Fatalf("shared skill missing in skills-first merge")
	}
	if sharedSkill.Description != "skill-shared" || sharedSkill.FilePath != "/skills/shared.md" {
		t.Fatalf("expected skill to win when precedence is PreferSkillFiles, got %#v", sharedSkill)
	}
	pluginOnlySkill, ok := findSkillByName(skillsFirst, "plugin-only")
	if !ok {
		t.Fatalf("plugin-only skill missing in skills-first merge")
	}
	if pluginOnlySkill.FilePath != "plugin:alpha" {
		t.Fatalf("expected plugin-derived filepath marker, got %#v", pluginOnlySkill)
	}

	pluginsFirst := MergeSkillsWithPluginCommands(base, pluginCommands, PreferPluginCommands)
	sharedPlugin, ok := findSkillByName(pluginsFirst, "shared")
	if !ok {
		t.Fatalf("shared skill missing in plugins-first merge")
	}
	if sharedPlugin.Description != "plugin-shared" || sharedPlugin.FilePath != "plugin:alpha" {
		t.Fatalf("expected plugin to win when precedence is PreferPluginCommands, got %#v", sharedPlugin)
	}
}

func TestMergeSkillsWithPluginCommandsDetailedDiagnosticsAndOrigins(t *testing.T) {
	base := []*Skill{
		{Name: "shared", Description: "skill-shared", FilePath: "/skills/shared.md", Enabled: true},
	}

	pluginCommands := []plugins.ResolvedCommand{
		{PluginID: "alpha", Name: "shared", Description: "plugin-shared", Usage: "/shared"},
		{PluginID: "alpha", Name: "plugin-only", Description: "plugin-only", Usage: "/plugin-only"},
	}

	detailed := MergeSkillsWithPluginCommandsDetailed(base, pluginCommands, PreferSkillFiles)
	if len(detailed.Skills) != 2 {
		t.Fatalf("expected two merged skills, got %#v", detailed.Skills)
	}
	if got := detailed.PluginOrigins["plugin-only"]; got != "alpha" {
		t.Fatalf("expected plugin origin for plugin-only, got %q", got)
	}
	if len(detailed.Conflicts) != 1 {
		t.Fatalf("expected one conflict diagnostic, got %#v", detailed.Conflicts)
	}
	if detailed.Conflicts[0].Name != "shared" || detailed.Conflicts[0].WinnerSource != "/skills/shared.md" {
		t.Fatalf("unexpected conflict diagnostic: %#v", detailed.Conflicts[0])
	}

	detailedPluginsFirst := MergeSkillsWithPluginCommandsDetailed(base, pluginCommands, PreferPluginCommands)
	if len(detailedPluginsFirst.Conflicts) != 1 {
		t.Fatalf("expected one conflict in plugins-first merge, got %#v", detailedPluginsFirst.Conflicts)
	}
	if detailedPluginsFirst.Conflicts[0].WinnerSource != "plugin:alpha" {
		t.Fatalf("expected plugin winner source, got %#v", detailedPluginsFirst.Conflicts[0])
	}
}

func findSkillByName(skills []*Skill, name string) (*Skill, bool) {
	for _, skill := range skills {
		if skill != nil && skill.Name == name {
			return skill, true
		}
	}
	return nil, false
}
