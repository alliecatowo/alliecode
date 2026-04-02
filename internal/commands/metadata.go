package commands

import "strings"

type CommandMetadata struct {
	Category     string
	Group        string
	PaletteGroup string
	ArgumentHint string
	HelpHint     string
	Shortcuts    []string
	Keywords     []string
	Examples     []string
	Contexts     []string
	Diagnostics  []string
}

func commandMetadataForName(name string) CommandMetadata {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return CommandMetadata{Category: "general", Group: "core", PaletteGroup: "core", HelpHint: "Start with /help <topic>"}
	}
	if meta, ok := commandMetadataTable[name]; ok {
		if strings.TrimSpace(meta.Group) == "" {
			meta.Group = "core"
		}
		if strings.TrimSpace(meta.PaletteGroup) == "" {
			meta.PaletteGroup = meta.Group
		}
		if strings.TrimSpace(meta.HelpHint) == "" {
			meta.HelpHint = "Run /help " + name + " for examples"
		}
		return meta
	}
	return CommandMetadata{
		Category:     inferCommandCategory(name),
		Group:        "core",
		PaletteGroup: "core",
		HelpHint:     "Run /help " + name + " for examples",
	}
}

var commandMetadataTable = composeCommandMetadataTable()

func composeCommandMetadataTable() map[string]CommandMetadata {
	table := map[string]CommandMetadata{}
	for _, family := range commandMetadataFamilies() {
		for name, meta := range family {
			table[name] = meta
		}
	}
	return table
}

func commandMetadataFamilies() []map[string]CommandMetadata {
	return []map[string]CommandMetadata{
		generalMetadata(),
		authMetadata(),
		accountMetadata(),
		configurationMetadata(),
		diagnosticsMetadata(),
		sessionMetadata(),
		contextMetadata(),
		gitMetadata(),
		workflowMetadata(),
		integrationsMetadata(),
		securityMetadata(),
	}
}

func inferCommandCategory(name string) string {
	if strings.Contains(name, "mcp") || strings.Contains(name, "plugin") || strings.Contains(name, "web") {
		return "integrations"
	}
	if strings.Contains(name, "auth") || strings.Contains(name, "login") || strings.Contains(name, "oauth") {
		return "auth"
	}
	if strings.Contains(name, "branch") || strings.Contains(name, "diff") || strings.Contains(name, "commit") || strings.Contains(name, "review") || strings.Contains(name, "pr") {
		return "git"
	}
	if strings.Contains(name, "session") || strings.Contains(name, "resume") || strings.Contains(name, "history") {
		return "session"
	}
	if strings.Contains(name, "config") || strings.Contains(name, "theme") || strings.Contains(name, "style") || strings.Contains(name, "terminal") {
		return "configuration"
	}
	if strings.Contains(name, "doctor") || strings.Contains(name, "status") || strings.Contains(name, "stats") {
		return "diagnostics"
	}
	if strings.Contains(name, "task") || strings.Contains(name, "workflow") || strings.Contains(name, "issue") || strings.Contains(name, "assistant") {
		return "workflow"
	}
	return "general"
}
