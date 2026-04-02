package plugins

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

var pluginIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

// Manifest describes one plugin and the metadata it contributes.
type Manifest struct {
	ID       string            `yaml:"id"`
	Name     string            `yaml:"name,omitempty"`
	Version  string            `yaml:"version"`
	Commands []CommandManifest `yaml:"commands,omitempty"`
	Tools    []ToolManifest    `yaml:"tools,omitempty"`
}

// CommandManifest is metadata for a slash command contributed by a plugin.
type CommandManifest struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
	Usage       string `yaml:"usage,omitempty"`
}

// ToolManifest is metadata for a tool contributed by a plugin.
type ToolManifest struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

// Policy controls which discovered plugins are active.
type Policy struct {
	Enabled  []string
	Disabled []string
}

// InstalledPlugin represents one discovered plugin directory.
type InstalledPlugin struct {
	Directory string
	Manifest  Manifest
	Enabled   bool
}

// ResolvedCommand is command metadata with plugin ownership attached.
type ResolvedCommand struct {
	PluginID    string
	PluginName  string
	Name        string
	Description string
	Usage       string
}

// ResolvedTool is tool metadata with plugin ownership attached.
type ResolvedTool struct {
	PluginID    string
	PluginName  string
	Name        string
	Description string
}

// Runtime is the resolved plugin state used by the application.
type Runtime struct {
	plugins  []InstalledPlugin
	commands []ResolvedCommand
	tools    []ResolvedTool
}

// Plugins lists all discovered plugins, including disabled ones.
func (r *Runtime) Plugins() []InstalledPlugin {
	out := make([]InstalledPlugin, len(r.plugins))
	copy(out, r.plugins)
	return out
}

// Commands lists metadata for commands from enabled plugins.
func (r *Runtime) Commands() []ResolvedCommand {
	out := make([]ResolvedCommand, len(r.commands))
	copy(out, r.commands)
	return out
}

// Tools lists metadata for tools from enabled plugins.
func (r *Runtime) Tools() []ResolvedTool {
	out := make([]ResolvedTool, len(r.tools))
	copy(out, r.tools)
	return out
}

// LoadRuntime discovers plugin manifests from immediate sub-directories.
func LoadRuntime(root string, policy Policy) (*Runtime, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read plugin root %q: %w", root, err)
	}

	type discovered struct {
		directory string
		manifest  Manifest
	}

	var (
		found   []discovered
		errList []error
	)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join(root, entry.Name())
		manifestPath, ok := findManifestPath(dir)
		if !ok {
			continue
		}

		manifest, err := loadManifest(manifestPath)
		if err != nil {
			errList = append(errList, fmt.Errorf("load manifest %q: %w", manifestPath, err))
			continue
		}
		normalizeManifest(&manifest)
		if err := ValidateManifest(manifest); err != nil {
			errList = append(errList, fmt.Errorf("invalid manifest %q: %w", manifestPath, err))
			continue
		}

		found = append(found, discovered{directory: dir, manifest: manifest})
	}

	if len(errList) > 0 {
		return nil, errors.Join(errList...)
	}

	sort.Slice(found, func(i, j int) bool {
		return found[i].manifest.ID < found[j].manifest.ID
	})

	ids := make([]string, 0, len(found))
	idIndex := make(map[string]string, len(found))
	for _, p := range found {
		if priorDir, exists := idIndex[p.manifest.ID]; exists {
			return nil, fmt.Errorf("plugin id %q is duplicated across %q and %q", p.manifest.ID, priorDir, p.directory)
		}
		idIndex[p.manifest.ID] = p.directory
		ids = append(ids, p.manifest.ID)
	}

	if err := policy.Validate(ids); err != nil {
		return nil, err
	}

	runtime := &Runtime{plugins: make([]InstalledPlugin, 0, len(found))}
	seenCommands := make(map[string]string)
	seenTools := make(map[string]string)

	for _, p := range found {
		enabled := policy.IsEnabled(p.manifest.ID)
		runtime.plugins = append(runtime.plugins, InstalledPlugin{
			Directory: p.directory,
			Manifest:  p.manifest,
			Enabled:   enabled,
		})

		if !enabled {
			continue
		}

		for _, cmd := range p.manifest.Commands {
			if prior, exists := seenCommands[cmd.Name]; exists {
				return nil, fmt.Errorf("command %q declared by both %q and %q", cmd.Name, prior, p.manifest.ID)
			}
			seenCommands[cmd.Name] = p.manifest.ID
			runtime.commands = append(runtime.commands, ResolvedCommand{
				PluginID:    p.manifest.ID,
				PluginName:  p.manifest.Name,
				Name:        cmd.Name,
				Description: cmd.Description,
				Usage:       cmd.Usage,
			})
		}

		for _, tool := range p.manifest.Tools {
			if prior, exists := seenTools[tool.Name]; exists {
				return nil, fmt.Errorf("tool %q declared by both %q and %q", tool.Name, prior, p.manifest.ID)
			}
			seenTools[tool.Name] = p.manifest.ID
			runtime.tools = append(runtime.tools, ResolvedTool{
				PluginID:    p.manifest.ID,
				PluginName:  p.manifest.Name,
				Name:        tool.Name,
				Description: tool.Description,
			})
		}
	}

	sort.Slice(runtime.commands, func(i, j int) bool {
		if runtime.commands[i].Name != runtime.commands[j].Name {
			return runtime.commands[i].Name < runtime.commands[j].Name
		}
		return runtime.commands[i].PluginID < runtime.commands[j].PluginID
	})

	sort.Slice(runtime.tools, func(i, j int) bool {
		if runtime.tools[i].Name != runtime.tools[j].Name {
			return runtime.tools[i].Name < runtime.tools[j].Name
		}
		return runtime.tools[i].PluginID < runtime.tools[j].PluginID
	})

	return runtime, nil
}

// ValidateManifest checks required fields and duplicate command/tool names.
func ValidateManifest(m Manifest) error {
	id := strings.TrimSpace(m.ID)
	if id == "" {
		return fmt.Errorf("id is required")
	}
	if !pluginIDPattern.MatchString(id) {
		return fmt.Errorf("id %q must match %s", id, pluginIDPattern.String())
	}
	if strings.TrimSpace(m.Version) == "" {
		return fmt.Errorf("version is required")
	}

	seenCommands := make(map[string]struct{}, len(m.Commands))
	for i, cmd := range m.Commands {
		name := strings.TrimSpace(cmd.Name)
		if name == "" {
			return fmt.Errorf("commands[%d].name is required", i)
		}
		if _, exists := seenCommands[name]; exists {
			return fmt.Errorf("duplicate command name %q", name)
		}
		seenCommands[name] = struct{}{}
	}

	seenTools := make(map[string]struct{}, len(m.Tools))
	for i, tool := range m.Tools {
		name := strings.TrimSpace(tool.Name)
		if name == "" {
			return fmt.Errorf("tools[%d].name is required", i)
		}
		if _, exists := seenTools[name]; exists {
			return fmt.Errorf("duplicate tool name %q", name)
		}
		seenTools[name] = struct{}{}
	}

	return nil
}

// Validate checks policy coherence and unknown IDs against discovered plugins.
func (p Policy) Validate(knownPluginIDs []string) error {
	known := make(map[string]struct{}, len(knownPluginIDs))
	for _, id := range knownPluginIDs {
		known[strings.TrimSpace(id)] = struct{}{}
	}

	seenEnabled := make(map[string]struct{}, len(p.Enabled))
	for _, id := range p.Enabled {
		id = strings.TrimSpace(id)
		if id == "" {
			return fmt.Errorf("enabled policy contains empty id")
		}
		if _, exists := seenEnabled[id]; exists {
			return fmt.Errorf("enabled policy contains duplicate id %q", id)
		}
		seenEnabled[id] = struct{}{}
		if _, exists := known[id]; !exists {
			return fmt.Errorf("enabled policy references unknown plugin %q", id)
		}
	}

	seenDisabled := make(map[string]struct{}, len(p.Disabled))
	for _, id := range p.Disabled {
		id = strings.TrimSpace(id)
		if id == "" {
			return fmt.Errorf("disabled policy contains empty id")
		}
		if _, exists := seenDisabled[id]; exists {
			return fmt.Errorf("disabled policy contains duplicate id %q", id)
		}
		seenDisabled[id] = struct{}{}
		if _, exists := known[id]; !exists {
			return fmt.Errorf("disabled policy references unknown plugin %q", id)
		}
		if _, exists := seenEnabled[id]; exists {
			return fmt.Errorf("plugin %q is listed in both enabled and disabled policies", id)
		}
	}

	return nil
}

// IsEnabled applies policy semantics for one plugin id.
func (p Policy) IsEnabled(pluginID string) bool {
	pluginID = strings.TrimSpace(pluginID)
	for _, id := range p.Disabled {
		if strings.TrimSpace(id) == pluginID {
			return false
		}
	}
	if len(p.Enabled) == 0 {
		return true
	}
	for _, id := range p.Enabled {
		if strings.TrimSpace(id) == pluginID {
			return true
		}
	}
	return false
}

func normalizeManifest(m *Manifest) {
	m.ID = strings.TrimSpace(m.ID)
	m.Name = strings.TrimSpace(m.Name)
	m.Version = strings.TrimSpace(m.Version)

	for i := range m.Commands {
		m.Commands[i].Name = strings.TrimSpace(m.Commands[i].Name)
		m.Commands[i].Description = strings.TrimSpace(m.Commands[i].Description)
		m.Commands[i].Usage = strings.TrimSpace(m.Commands[i].Usage)
	}
	for i := range m.Tools {
		m.Tools[i].Name = strings.TrimSpace(m.Tools[i].Name)
		m.Tools[i].Description = strings.TrimSpace(m.Tools[i].Description)
	}
}

func findManifestPath(pluginDir string) (string, bool) {
	candidates := []string{"plugin.yaml", "plugin.yml", "plugin.json"}
	for _, candidate := range candidates {
		path := filepath.Join(pluginDir, candidate)
		if st, err := os.Stat(path); err == nil && !st.IsDir() {
			return path, true
		}
	}
	return "", false
}

func loadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}

	var m Manifest
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&m); err != nil {
		return Manifest{}, err
	}
	return m, nil
}
