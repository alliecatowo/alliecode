// Package skills implements the skill/slash-command system for AllieCode.
// Skills are reusable prompt templates with tool requirements, defined as markdown files.
package skills

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alliecatowo/alliecode/internal/plugins"
	"gopkg.in/yaml.v3"
)

// Skill represents a reusable prompt template with tool requirements.
type Skill struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Tags        []string `yaml:"tags"`
	Tools       []string `yaml:"tools"`
	Enabled     bool     `yaml:"enabled"`
	Prompt      string   `yaml:"-"` // the markdown body after frontmatter
	FilePath    string   `yaml:"-"` // where loaded from
	Source      string   `yaml:"-"`
	PluginID    string   `yaml:"-"`
}

// FilterOptions controls skill listing filters.
type FilterOptions struct {
	NameContains    string
	ExactName       string
	Tag             string
	Tool            string
	IncludeDisabled bool
}

// Manager manages the lifecycle of skills.
type Manager struct {
	dirs           []string
	discoveredDirs map[string]struct{}
	skills         map[string]*Skill
}

// MergePrecedence controls winner selection when names collide.
type MergePrecedence int

const (
	// PreferSkillFiles keeps file-discovered skills when a plugin contributes
	// the same name.
	PreferSkillFiles MergePrecedence = iota
	// PreferPluginCommands lets plugin-contributed commands override
	// file-discovered skills for the same name.
	PreferPluginCommands
)

// SkillConflictDiagnostic reports one file/plugin naming conflict in merged listings.
type SkillConflictDiagnostic struct {
	Name           string `json:"name"`
	ExistingSource string `json:"existing_source"`
	IncomingSource string `json:"incoming_source"`
	WinnerSource   string `json:"winner_source"`
}

// SkillMetadata describes deterministic provenance for one merged skill.
type SkillMetadata struct {
	Name         string `json:"name"`
	Source       string `json:"source"`
	FilePath     string `json:"file_path,omitempty"`
	PluginID     string `json:"plugin_id,omitempty"`
	WinnerSource string `json:"winner_source,omitempty"`
}

// MergedSkillList captures merged skill output with conflict diagnostics.
type MergedSkillList struct {
	Skills        []*Skill                  `json:"skills"`
	PluginOrigins map[string]string         `json:"plugin_origins,omitempty"`
	Metadata      map[string]SkillMetadata  `json:"metadata,omitempty"`
	PluginOverlay map[string][]string       `json:"plugin_overlay,omitempty"`
	Conflicts     []SkillConflictDiagnostic `json:"conflicts,omitempty"`
}

// NewManager creates a new skill manager that scans the given directories.
func NewManager(dirs []string) *Manager {
	if len(dirs) == 0 {
		dirs = DefaultSkillDirs("")
	}

	return &Manager{
		dirs:           dirs,
		discoveredDirs: make(map[string]struct{}),
		skills:         make(map[string]*Skill),
	}
}

// DefaultSkillDirs returns workspace and global skill directories.
// If workspaceDir is empty, the current working directory is used.
func DefaultSkillDirs(workspaceDir string) []string {
	if workspaceDir == "" {
		wd, err := os.Getwd()
		if err == nil {
			workspaceDir = wd
		}
	}

	dirs := make([]string, 0, 2)
	if workspaceDir != "" {
		dirs = append(dirs, filepath.Join(workspaceDir, ".alliecode", "skills"))
	}

	home, err := os.UserHomeDir()
	if err == nil && home != "" {
		global := filepath.Join(home, ".alliecode", "skills")
		if len(dirs) == 0 || dirs[0] != global {
			dirs = append(dirs, global)
		}
	}

	return dirs
}

// SkillDirs returns configured and discovered skill directories.
func (m *Manager) SkillDirs() []string {
	out := make([]string, len(m.dirs))
	copy(out, m.dirs)
	return out
}

// Load scans all configured directories for skill definition files (.md).
func (m *Manager) Load() error {
	m.skills = make(map[string]*Skill)

	for _, dir := range m.dirs {
		files, err := discoverSkillFiles(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue // directory doesn't exist yet, that's fine
			}
			return fmt.Errorf("reading skills dir %q: %w", dir, err)
		}

		for _, path := range files {
			skill, parseErr := parseSkillFile(path)
			if parseErr != nil {
				continue // skip invalid skills
			}

			if _, exists := m.skills[skill.Name]; exists {
				continue
			}

			m.skills[skill.Name] = skill
		}
	}

	return nil
}

// DiscoverSkillDirsForPaths discovers nested .alliecode/skills directories for file paths.
// Returned directories are sorted deepest first to keep nearest-path precedence stable.
func (m *Manager) DiscoverSkillDirsForPaths(filePaths []string, cwd string) ([]string, error) {
	resolvedCwd := strings.TrimSuffix(cwd, string(filepath.Separator))
	if strings.TrimSpace(resolvedCwd) == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		resolvedCwd = wd
	}

	found := make([]string, 0)
	seen := make(map[string]struct{})

	for _, filePath := range filePaths {
		if strings.TrimSpace(filePath) == "" {
			continue
		}
		current := filepath.Dir(filePath)
		for strings.HasPrefix(current, resolvedCwd+string(filepath.Separator)) {
			skillDir := filepath.Join(current, ".alliecode", "skills")
			if _, alreadyKnown := m.discoveredDirs[skillDir]; alreadyKnown {
				parent := filepath.Dir(current)
				if parent == current {
					break
				}
				current = parent
				continue
			}
			if _, seenNow := seen[skillDir]; !seenNow {
				if st, err := os.Stat(skillDir); err == nil && st.IsDir() {
					seen[skillDir] = struct{}{}
					found = append(found, skillDir)
				}
			}
			parent := filepath.Dir(current)
			if parent == current {
				break
			}
			current = parent
		}
	}

	sort.Slice(found, func(i, j int) bool {
		li := strings.Count(found[i], string(filepath.Separator))
		lj := strings.Count(found[j], string(filepath.Separator))
		if li != lj {
			return li > lj
		}
		return found[i] < found[j]
	})
	return found, nil
}

// AddDiscoveredDirectories loads skills from discovered directories and merges them.
func (m *Manager) AddDiscoveredDirectories(dirs []string) error {
	for _, dir := range dirs {
		trimmed := strings.TrimSpace(dir)
		if trimmed == "" {
			continue
		}
		if _, exists := m.discoveredDirs[trimmed]; exists {
			continue
		}
		m.discoveredDirs[trimmed] = struct{}{}
		m.dirs = append(m.dirs, trimmed)
	}
	return m.Load()
}

func discoverSkillFiles(dir string) ([]string, error) {
	files := make([]string, 0)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		if d.IsDir() || !strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// Get returns a skill by name.
func (m *Manager) Get(name string) (*Skill, bool) {
	skill, ok := m.skills[name]
	return skill, ok
}

// List returns all loaded skills.
func (m *Manager) List() []*Skill {
	return m.Filter(FilterOptions{IncludeDisabled: true})
}

// ListMerged returns file-discovered skills merged with plugin commands.
// Precedence decides which source wins on name collisions.
func (m *Manager) ListMerged(pluginCommands []plugins.ResolvedCommand, precedence MergePrecedence) []*Skill {
	base := m.List()
	return MergeSkillsWithPluginCommandsDetailed(base, pluginCommands, precedence).Skills
}

// MergeSkillsWithPluginCommands combines skill files and plugin commands into
// one deterministic listing with explicit collision precedence.
func MergeSkillsWithPluginCommands(
	skills []*Skill,
	pluginCommands []plugins.ResolvedCommand,
	precedence MergePrecedence,
) []*Skill {
	return MergeSkillsWithPluginCommandsDetailed(skills, pluginCommands, precedence).Skills
}

// MergeSkillsWithPluginCommandsDetailed combines skill files and plugin
// commands while preserving deterministic conflict diagnostics and plugin
// origin data for downstream list UIs and tool outputs.
func MergeSkillsWithPluginCommandsDetailed(
	skills []*Skill,
	pluginCommands []plugins.ResolvedCommand,
	precedence MergePrecedence,
) MergedSkillList {
	merged := make(map[string]*Skill, len(skills)+len(pluginCommands))
	pluginOrigins := make(map[string]string, len(pluginCommands))
	metadata := make(map[string]SkillMetadata, len(skills)+len(pluginCommands))
	pluginOverlay := make(map[string][]string, len(pluginCommands))
	conflicts := make([]SkillConflictDiagnostic, 0)

	for _, s := range skills {
		if s == nil {
			continue
		}
		copy := *s
		if strings.TrimSpace(copy.Source) == "" {
			copy.Source = "file"
		}
		merged[s.Name] = &copy
		metadata[s.Name] = SkillMetadata{
			Name:     s.Name,
			Source:   copy.Source,
			FilePath: copy.FilePath,
			PluginID: copy.PluginID,
		}
	}

	for _, cmd := range pluginCommands {
		name := strings.TrimSpace(cmd.Name)
		if name == "" {
			continue
		}

		existing, exists := merged[name]
		if exists && precedence == PreferSkillFiles {
			incoming := fmt.Sprintf("plugin:%s", cmd.PluginID)
			pluginOverlay[name] = appendUniqueString(pluginOverlay[name], cmd.PluginID)
			conflicts = append(conflicts, SkillConflictDiagnostic{
				Name:           name,
				ExistingSource: existing.FilePath,
				IncomingSource: incoming,
				WinnerSource:   existing.FilePath,
			})
			meta := metadata[name]
			meta.Name = name
			if meta.Source == "" {
				meta.Source = "file"
			}
			meta.WinnerSource = existing.FilePath
			metadata[name] = meta
			continue
		}
		if exists {
			incoming := fmt.Sprintf("plugin:%s", cmd.PluginID)
			conflicts = append(conflicts, SkillConflictDiagnostic{
				Name:           name,
				ExistingSource: existing.FilePath,
				IncomingSource: incoming,
				WinnerSource:   incoming,
			})
			pluginOverlay[name] = appendUniqueString(pluginOverlay[name], cmd.PluginID)
		}

		desc := strings.TrimSpace(cmd.Description)
		if desc == "" {
			desc = fmt.Sprintf("Plugin command from %s", cmd.PluginID)
		}

		merged[name] = &Skill{
			Name:        name,
			Description: desc,
			Tags:        []string{"plugin", fmt.Sprintf("plugin:%s", cmd.PluginID)},
			Tools:       nil,
			Enabled:     true,
			Prompt:      strings.TrimSpace(cmd.Usage),
			FilePath:    fmt.Sprintf("plugin:%s", cmd.PluginID),
			Source:      "plugin",
			PluginID:    cmd.PluginID,
		}
		pluginOrigins[name] = cmd.PluginID
		metadata[name] = SkillMetadata{
			Name:         name,
			Source:       "plugin",
			FilePath:     fmt.Sprintf("plugin:%s", cmd.PluginID),
			PluginID:     cmd.PluginID,
			WinnerSource: fmt.Sprintf("plugin:%s", cmd.PluginID),
		}
		pluginOverlay[name] = appendUniqueString(pluginOverlay[name], cmd.PluginID)
	}

	out := make([]*Skill, 0, len(merged))
	for _, s := range merged {
		out = append(out, s)
	}

	sort.Slice(out, func(i, j int) bool {
		left := strings.ToLower(out[i].Name)
		right := strings.ToLower(out[j].Name)
		if left != right {
			return left < right
		}
		return out[i].FilePath < out[j].FilePath
	})

	sort.Slice(conflicts, func(i, j int) bool {
		if strings.ToLower(conflicts[i].Name) != strings.ToLower(conflicts[j].Name) {
			return strings.ToLower(conflicts[i].Name) < strings.ToLower(conflicts[j].Name)
		}
		if conflicts[i].WinnerSource != conflicts[j].WinnerSource {
			return conflicts[i].WinnerSource < conflicts[j].WinnerSource
		}
		return conflicts[i].IncomingSource < conflicts[j].IncomingSource
	})

	for name, meta := range metadata {
		if meta.WinnerSource == "" {
			if skill, ok := merged[name]; ok {
				meta.WinnerSource = skill.FilePath
			}
			metadata[name] = meta
		}
	}

	return MergedSkillList{Skills: out, PluginOrigins: pluginOrigins, Metadata: metadata, PluginOverlay: pluginOverlay, Conflicts: conflicts}
}

func appendUniqueString(list []string, value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return list
	}
	for _, existing := range list {
		if existing == trimmed {
			return list
		}
	}
	return append(list, trimmed)
}

// Filter returns loaded skills matching filter criteria.
func (m *Manager) Filter(opts FilterOptions) []*Skill {
	skills := make([]*Skill, 0, len(m.skills))
	for _, s := range m.skills {
		if !opts.IncludeDisabled && !s.Enabled {
			continue
		}
		if opts.ExactName != "" && !strings.EqualFold(strings.TrimSpace(s.Name), strings.TrimSpace(opts.ExactName)) {
			continue
		}
		if opts.NameContains != "" && !strings.Contains(strings.ToLower(s.Name), strings.ToLower(strings.TrimSpace(opts.NameContains))) {
			continue
		}
		if opts.Tag != "" && !hasTag(s.Tags, opts.Tag) {
			continue
		}
		if opts.Tool != "" && !hasTag(s.Tools, opts.Tool) {
			continue
		}
		skills = append(skills, s)
	}

	sort.Slice(skills, func(i, j int) bool {
		left := strings.ToLower(skills[i].Name)
		right := strings.ToLower(skills[j].Name)
		if left != right {
			return left < right
		}
		return skills[i].FilePath < skills[j].FilePath
	})

	return skills
}

// Execute returns the expanded prompt for a skill with the given arguments.
func (m *Manager) Execute(name string, args string) (string, error) {
	skill, ok := m.skills[name]
	if !ok {
		return "", fmt.Errorf("skill %q not found", name)
	}
	if !skill.Enabled {
		return "", fmt.Errorf("skill %q is disabled", name)
	}

	prompt := skill.Prompt
	if args != "" {
		prompt = fmt.Sprintf("%s\n\nArguments: %s", prompt, args)
	}

	return prompt, nil
}

type skillMetadata struct {
	Name        string   `yaml:"name"`
	Title       string   `yaml:"title"`
	Description string   `yaml:"description"`
	Summary     string   `yaml:"summary"`
	Tags        yamlList `yaml:"tags"`
	Tag         yamlList `yaml:"tag"`
	Tools       yamlList `yaml:"tools"`
	Tool        yamlList `yaml:"tool"`
	Enabled     *bool    `yaml:"enabled"`
}

type yamlList []string

func (l *yamlList) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		if strings.TrimSpace(value.Value) == "" {
			*l = nil
			return nil
		}
		*l = []string{value.Value}
		return nil
	case yaml.SequenceNode:
		items := make([]string, 0, len(value.Content))
		for _, node := range value.Content {
			if node.Kind != yaml.ScalarNode {
				return fmt.Errorf("list values must be scalar")
			}
			items = append(items, node.Value)
		}
		*l = items
		return nil
	default:
		return fmt.Errorf("expected scalar or sequence")
	}
}

// parseSkillFile reads a markdown file with YAML frontmatter and extracts the skill definition.
func parseSkillFile(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(data)
	meta := skillMetadata{}
	body := strings.TrimSpace(content)

	if strings.HasPrefix(content, "---\n") {
		parsedMeta, parsedBody, parseErr := parseFrontmatter(content)
		if parseErr != nil {
			return nil, parseErr
		}
		meta = parsedMeta
		body = parsedBody
	} else {
		parsedMeta, parsedBody, ok := parseHeadingMetadata(content)
		if ok {
			meta = parsedMeta
			body = parsedBody
		}
	}

	enabled := true
	if meta.Enabled != nil {
		enabled = *meta.Enabled
	}

	name := strings.TrimSpace(meta.Name)
	if name == "" {
		name = strings.TrimSpace(meta.Title)
	}
	if name == "" {
		name = strings.TrimSuffix(filepath.Base(path), ".md")
	}

	description := strings.TrimSpace(meta.Description)
	if description == "" {
		description = strings.TrimSpace(meta.Summary)
	}

	tags := append([]string{}, []string(meta.Tags)...)
	tags = append(tags, []string(meta.Tag)...)

	tools := append([]string{}, []string(meta.Tools)...)
	tools = append(tools, []string(meta.Tool)...)

	skill := Skill{
		Name:        name,
		Description: description,
		Tags:        normalizeList(tags),
		Tools:       normalizeList(tools),
		Enabled:     enabled,
		Prompt:      body,
		FilePath:    path,
		Source:      "file",
	}

	return &skill, nil
}

func parseFrontmatter(content string) (skillMetadata, string, error) {
	endIndex := strings.Index(content[4:], "\n---")
	if endIndex == -1 {
		return skillMetadata{}, "", fmt.Errorf("skill file has unclosed frontmatter")
	}

	frontmatter := content[4 : 4+endIndex]
	body := strings.TrimSpace(content[4+endIndex+4:])

	var meta skillMetadata
	if err := yaml.Unmarshal([]byte(frontmatter), &meta); err != nil {
		return skillMetadata{}, "", fmt.Errorf("parsing skill frontmatter: %w", err)
	}

	return meta, body, nil
}

func parseHeadingMetadata(content string) (skillMetadata, string, bool) {
	lines := strings.Split(content, "\n")
	i := 0
	for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
		i++
	}
	if i >= len(lines) {
		return skillMetadata{}, "", false
	}

	line := strings.TrimSpace(lines[i])
	if !strings.HasPrefix(line, "#") {
		return skillMetadata{}, strings.TrimSpace(content), false
	}

	name := strings.TrimSpace(strings.TrimLeft(line, "#"))
	i++

	meta := skillMetadata{Name: name}
	for i < len(lines) {
		current := strings.TrimSpace(lines[i])
		if current == "" {
			i++
			break
		}

		key, value, ok := strings.Cut(current, ":")
		if !ok {
			break
		}

		switch strings.ToLower(strings.TrimSpace(key)) {
		case "name":
			meta.Name = strings.TrimSpace(value)
		case "title":
			meta.Title = strings.TrimSpace(value)
		case "description":
			meta.Description = strings.TrimSpace(value)
		case "summary":
			meta.Summary = strings.TrimSpace(value)
		case "tags":
			meta.Tags = yamlList(parseListValue(value))
		case "tag":
			meta.Tag = yamlList(parseListValue(value))
		case "tools":
			meta.Tools = yamlList(parseListValue(value))
		case "tool":
			meta.Tool = yamlList(parseListValue(value))
		case "enabled":
			if parsed, err := parseBool(value); err == nil {
				meta.Enabled = &parsed
			}
		default:
			break
		}
		i++
	}

	body := strings.TrimSpace(strings.Join(lines[i:], "\n"))
	return meta, body, true
}

func parseListValue(raw string) []string {
	value := strings.TrimSpace(raw)
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.Trim(strings.TrimSpace(part), "\"'")
		if item == "" {
			continue
		}
		items = append(items, item)
	}
	return items
}

func parseBool(raw string) (bool, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "true", "yes", "on", "1":
		return true, nil
	case "false", "no", "off", "0":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool")
	}
}

func normalizeList(items []string) []string {
	normalized := make([]string, 0, len(items))
	seen := make(map[string]struct{})
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func hasTag(tags []string, want string) bool {
	target := strings.ToLower(strings.TrimSpace(want))
	for _, tag := range tags {
		if strings.ToLower(tag) == target {
			return true
		}
	}
	return false
}
