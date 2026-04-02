package commands

func configurationMetadata() map[string]CommandMetadata {
	return map[string]CommandMetadata{
		"config":         {Category: "configuration", Group: "runtime", PaletteGroup: "startup", ArgumentHint: "[show|status|doctor|panel|get|set|unset|repair]", HelpHint: "Use /config doctor, then /config repair <profile>", Shortcuts: []string{"settings"}, Keywords: []string{"settings", "profiles", "repair"}, Examples: []string{"/config status", "/config doctor", "/config repair safe"}, Contexts: []string{"startup", "runtime"}, Diagnostics: []string{"doctor", "repair", "status"}},
		"model":          {Category: "configuration", Group: "runtime", PaletteGroup: "startup", ArgumentHint: "[provider/model|model|list [provider]|doctor|repair [provider/model]]", HelpHint: "Use /model doctor for correction hints", Shortcuts: []string{"m"}, Keywords: []string{"switch", "llm", "capabilities"}, Examples: []string{"/model", "/model openai/gpt-4o", "/model doctor"}, Contexts: []string{"startup", "runtime-selection"}, Diagnostics: []string{"doctor", "repair"}},
		"provider":       {Category: "configuration", Group: "runtime", PaletteGroup: "startup", ArgumentHint: "[status|list|set <name>|doctor|repair [name]|models [name]]", HelpHint: "Use /provider doctor to validate credentials and model pair", Shortcuts: []string{"p"}, Keywords: []string{"auth", "backend", "credentials"}, Examples: []string{"/provider status", "/provider set ollama", "/provider doctor"}, Contexts: []string{"startup", "runtime-selection"}, Diagnostics: []string{"doctor", "repair", "models"}},
		"keybindings":    {Category: "configuration", Group: "ui", ArgumentHint: "[status|path|open|enable|disable]", HelpHint: "Enable first, then /keybindings open", Shortcuts: []string{"ctrl-k"}, Keywords: []string{"shortcuts", "keys"}, Examples: []string{"/keybindings status"}},
		"terminal-setup": {Category: "configuration", Group: "runtime", ArgumentHint: "[status|detect|apply]", HelpHint: "Detect profile before applying setup", Shortcuts: []string{"term"}, Keywords: []string{"terminal", "profile"}, Examples: []string{"/terminal-setup detect"}},
		"theme":          {Category: "configuration", Group: "ui", ArgumentHint: "[status|list|set|cycle|preview]", HelpHint: "Preview before set when scripting", Shortcuts: []string{"t"}, Keywords: []string{"colors", "appearance"}, Examples: []string{"/theme cycle"}},
		"output-style":   {Category: "configuration", Group: "ui", ArgumentHint: "[status|list|set]", HelpHint: "Use concise for compact terminal output", Shortcuts: []string{"style"}, Keywords: []string{"format", "verbosity"}, Examples: []string{"/output-style set concise"}},
		"remote-env":     {Category: "configuration", Group: "runtime", ArgumentHint: "[status|list|set <name>]", HelpHint: "Environment tags help reproducible runs", Shortcuts: []string{"env"}, Keywords: []string{"environment", "remote"}, Examples: []string{"/remote-env set hardened-linux"}},
	}
}

func diagnosticsMetadata() map[string]CommandMetadata {
	return map[string]CommandMetadata{
		"doctor": {Category: "diagnostics", Group: "health", PaletteGroup: "diagnostics", ArgumentHint: "[human|json|fix|diagnostics]", HelpHint: "Run /doctor fix for auto-correction guidance", Shortcuts: []string{"d"}, Keywords: []string{"health", "checks", "diagnostics"}, Examples: []string{"/doctor", "/doctor json", "/doctor fix"}, Contexts: []string{"runtime", "startup-repair", "corrective-loops"}, Diagnostics: []string{"provider", "model", "permissions", "settings", "history", "session", "mcp"}},
		"status": {Category: "diagnostics", Group: "health", PaletteGroup: "diagnostics", ArgumentHint: "[diagnostics]", HelpHint: "Shows runtime counters and startup metadata", Shortcuts: []string{"s"}, Keywords: []string{"runtime", "state", "health"}, Examples: []string{"/status", "/status diagnostics"}, Contexts: []string{"runtime", "session"}, Diagnostics: []string{"provider-loop", "model-loop", "permissions-loop", "settings-loop"}},
		"stats":  {Category: "diagnostics", Group: "health", ArgumentHint: "", HelpHint: "Use for lightweight counter snapshots", Shortcuts: []string{"metrics"}, Keywords: []string{"counters", "metrics"}, Examples: []string{"/stats"}},
	}
}
