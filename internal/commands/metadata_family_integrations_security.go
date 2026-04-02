package commands

func integrationsMetadata() map[string]CommandMetadata {
	return map[string]CommandMetadata{
		"mcp":    {Category: "integrations", Group: "extensions", PaletteGroup: "integrations", ArgumentHint: "[list|status|doctor|diagnostics|repair|enable|disable|reconnect|add|remove|connect|disconnect|list-tools|list-resources]", HelpHint: "Run /mcp doctor after server changes", Shortcuts: []string{"mc"}, Keywords: []string{"servers", "tools", "resources"}, Examples: []string{"/mcp list", "/mcp enable", "/mcp reconnect local"}, Contexts: []string{"extensions", "remote-tools"}, Diagnostics: []string{"doctor", "diagnostics", "repair"}},
		"plugin": {Category: "integrations", Group: "extensions", PaletteGroup: "integrations", ArgumentHint: "[list|status|doctor|diagnostics|repair|install|update|remove|enable|disable|marketplace]", HelpHint: "Run /plugin doctor before reload", Shortcuts: []string{"pl"}, Keywords: []string{"extensions", "marketplace"}, Examples: []string{"/plugin list", "/plugin doctor", "/plugin install acme/plugin@1.0.0"}, Contexts: []string{"extensions", "marketplace"}, Diagnostics: []string{"doctor", "diagnostics", "repair"}},
		"skills": {Category: "integrations", Group: "extensions", PaletteGroup: "integrations", ArgumentHint: "[list|status|doctor|repair|add|remove|sync]", HelpHint: "Use /skills sync to normalize from plugins", Shortcuts: []string{"sk"}, Keywords: []string{"skill", "capability"}, Examples: []string{"/skills list", "/skills sync"}, Contexts: []string{"skills", "prompts"}, Diagnostics: []string{"doctor", "repair", "sync"}},
		"ide":    {Category: "integrations", Group: "editor", ArgumentHint: "[status|detect|open|set-editor|hints]", HelpHint: "Detect once, then set-editor for deterministic opens", Shortcuts: []string{"ed"}, Keywords: []string{"editor", "vscode"}, Examples: []string{"/ide detect", "/ide open ."}},
	}
}

func securityMetadata() map[string]CommandMetadata {
	return map[string]CommandMetadata{
		"permissions":     {Category: "security", Group: "guardrails", PaletteGroup: "safety", ArgumentHint: "[status|set <mode>|rules ...|summary|denials|retry-denials]", HelpHint: "Use rules add/remove for fast policy tuning", Shortcuts: []string{"perm", "allowed-tools"}, Keywords: []string{"policy", "allow", "deny"}, Examples: []string{"/permissions status", "/permissions retry-denials"}, Contexts: []string{"safety", "sandbox"}, Diagnostics: []string{"summary", "denials", "retry-denials"}},
		"security-review": {Category: "security", Group: "guardrails", ArgumentHint: "[status|<target>]", HelpHint: "Run before release cut", Shortcuts: []string{"sec"}, Keywords: []string{"audit", "risk"}, Examples: []string{"/security-review"}},
	}
}
