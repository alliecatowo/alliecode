package commands

func integrationsMetadata() map[string]CommandMetadata {
	return map[string]CommandMetadata{
		"mcp":                {Category: "integrations", Group: "extensions", PaletteGroup: "integrations", ArgumentHint: "[list|status|doctor|diagnostics|repair|enable|disable|reconnect|add|remove|connect|disconnect|list-tools|list-resources]", HelpHint: "Run /mcp doctor then /mcp repair when server state drifts", Shortcuts: []string{"mc"}, Keywords: []string{"servers", "tools", "resources", "diagnostics"}, Examples: []string{"/mcp list", "/mcp doctor", "/mcp repair auto"}, Contexts: []string{"extensions", "remote-tools"}, Diagnostics: []string{"doctor", "diagnostics", "repair"}},
		"plugin":             {Category: "integrations", Group: "extensions", PaletteGroup: "integrations", ArgumentHint: "[list|status|doctor|diagnostics|repair|install|update|remove|enable|disable|marketplace]", HelpHint: "Run /plugin doctor, then /plugin repair or /reload-plugins", Shortcuts: []string{"pl"}, Keywords: []string{"extensions", "marketplace", "reload"}, Examples: []string{"/plugin status", "/plugin doctor", "/reload-plugins"}, Contexts: []string{"extensions", "marketplace"}, Diagnostics: []string{"doctor", "diagnostics", "repair"}},
		"skills":             {Category: "integrations", Group: "extensions", PaletteGroup: "integrations", ArgumentHint: "[list|status|doctor|repair [sync|auto|dedupe]|add <name>|remove <name>|sync]", HelpHint: "Use /skills list for source diagnostics, then /skills sync to resolve drift", Shortcuts: []string{"sk"}, Keywords: []string{"skill", "capability", "diagnostics", "plugins"}, Examples: []string{"/skills list", "/skills doctor", "/skills sync"}, Contexts: []string{"skills", "prompts"}, Diagnostics: []string{"doctor", "repair", "sync"}},
		"ide":                {Category: "integrations", Group: "editor", ArgumentHint: "[status|detect|open|set-editor|hints]", HelpHint: "Detect once, then set-editor for deterministic opens", Shortcuts: []string{"ed"}, Keywords: []string{"editor", "vscode"}, Examples: []string{"/ide detect", "/ide open ."}},
		"install-github-app": {Category: "integrations", Group: "extensions", ArgumentHint: "[status|<repo>]", HelpHint: "Install GitHub app integration", Keywords: []string{"github", "app"}, Examples: []string{"/install-github-app status"}},
		"install-slack-app":  {Category: "integrations", Group: "extensions", ArgumentHint: "[status|workspace <name>|install]", HelpHint: "Install Slack app integration", Keywords: []string{"slack", "app"}, Examples: []string{"/install-slack-app status"}},
		"web-setup":          {Category: "integrations", Group: "extensions", ArgumentHint: "[status|connect|disconnect]", HelpHint: "Link local session to web surface", Keywords: []string{"web", "remote"}, Examples: []string{"/web-setup connect"}},
		"reload-plugins":     {Category: "integrations", Group: "extensions", ArgumentHint: "[status|run]", HelpHint: "Reload plugin registry and state", Keywords: []string{"plugins", "reload"}, Examples: []string{"/reload-plugins run"}},
		"ant-trace":          {Category: "integrations", Group: "extensions", ArgumentHint: "[status|on|off|mark|list|clear]", HelpHint: "Track ant-only diagnostics traces", Keywords: []string{"trace", "diagnostics"}, Examples: []string{"/ant-trace mark startup"}},
		"debug-tool-call":    {Category: "integrations", Group: "extensions", ArgumentHint: "[status|log <tool>|clear]", HelpHint: "Track tool-call debug markers", Keywords: []string{"tools", "debug"}, Examples: []string{"/debug-tool-call log bash"}},
	}
}

func securityMetadata() map[string]CommandMetadata {
	return map[string]CommandMetadata{
		"permissions":     {Category: "security", Group: "guardrails", PaletteGroup: "safety", ArgumentHint: "[status|set <mode>|rules ...|summary|denials|retry-denials]", HelpHint: "Use rules add/remove for fast policy tuning", Shortcuts: []string{"perm", "allowed-tools"}, Keywords: []string{"policy", "allow", "deny"}, Examples: []string{"/permissions status", "/permissions retry-denials"}, Contexts: []string{"safety", "sandbox"}, Diagnostics: []string{"summary", "denials", "retry-denials"}},
		"security-review": {Category: "security", Group: "guardrails", ArgumentHint: "[status|<target>]", HelpHint: "Run before release cut", Shortcuts: []string{"sec"}, Keywords: []string{"audit", "risk"}, Examples: []string{"/security-review"}},
		"break-cache":     {Category: "security", Group: "guardrails", ArgumentHint: "[status|all|models|history|tools]", HelpHint: "Invalidate cache scopes when state drifts", Keywords: []string{"cache", "invalidate"}, Examples: []string{"/break-cache all"}},
		"bughunter":       {Category: "security", Group: "guardrails", ArgumentHint: "[status|run [scope]]", HelpHint: "Run deterministic bug reproduction sweeps", Keywords: []string{"bugs", "triage"}, Examples: []string{"/bughunter run auth"}},
	}
}
