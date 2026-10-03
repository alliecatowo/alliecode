package commands

func sessionMetadata() map[string]CommandMetadata {
	return map[string]CommandMetadata{
		"history":        {Category: "session", Group: "timeline", PaletteGroup: "context", ArgumentHint: "[list [model|text|limit|latest]|latest|show <id>|status|doctor]", HelpHint: "Use /history latest for quickest resume target", Shortcuts: []string{"h"}, Keywords: []string{"sessions", "resume", "timeline"}, Examples: []string{"/history list latest", "/history latest", "/history show session-1"}, Contexts: []string{"timeline", "resume"}, Diagnostics: []string{"status", "doctor"}},
		"session":        {Category: "session", Group: "remote", PaletteGroup: "context", ArgumentHint: "[status|set-url|host|connect|token|disconnect|diagnostics|repair]", HelpHint: "Use /session status after host/connect", Shortcuts: []string{"remote"}, Keywords: []string{"remote", "qr", "share"}, Examples: []string{"/session status", "/remote", "/session token set <token>"}, Contexts: []string{"remote", "handoff"}, Diagnostics: []string{"diagnostics", "repair"}},
		"resume":         {Category: "session", Group: "timeline", ArgumentHint: "[latest|status|<target>]", HelpHint: "Run /history list first when unsure", Shortcuts: []string{"continue"}, Keywords: []string{"continue", "restore"}, Examples: []string{"/resume", "/resume latest"}},
		"rewind":         {Category: "session", Group: "timeline", ArgumentHint: "[status|<target>]", HelpHint: "Use descriptive labels to make rewind safer", Shortcuts: []string{"checkpoint"}, Keywords: []string{"checkpoint", "rollback"}, Examples: []string{"/rewind status", "/rewind before-refactor"}},
		"rename":         {Category: "session", Group: "timeline", ArgumentHint: "[status|<title>]", HelpHint: "Session title persists into history records", Shortcuts: []string{"title"}, Keywords: []string{"title", "label"}, Examples: []string{"/rename status", "/rename onboarding fix"}},
		"tag":            {Category: "session", Group: "timeline", ArgumentHint: "[status|<tag>]", HelpHint: "Use tags to filter history quickly", Shortcuts: []string{"label"}, Keywords: []string{"labels", "search"}, Examples: []string{"/tag bugfix"}},
		"share":          {Category: "session", Group: "remote", ArgumentHint: "[status|list|create|revoke]", HelpHint: "Prefer private links unless review requires public", Shortcuts: []string{"publish"}, Keywords: []string{"publish", "links"}, Examples: []string{"/share create session public"}},
		"teleport":       {Category: "session", Group: "remote", ArgumentHint: "[status|<target>]", HelpHint: "Switch deterministic remote execution target", Keywords: []string{"remote", "handoff"}, Examples: []string{"/teleport laptop"}},
		"think-back":     {Category: "session", Group: "timeline", ArgumentHint: "[status|capture|restore]", HelpHint: "Capture checkpoints for future replay", Keywords: []string{"checkpoint", "history"}, Examples: []string{"/think-back capture"}},
		"thinkback-play": {Category: "session", Group: "timeline", ArgumentHint: "[status|play|stop]", HelpHint: "Replay deterministic thinkback checkpoints", Keywords: []string{"replay", "history"}, Examples: []string{"/thinkback-play play"}},
	}
}

func contextMetadata() map[string]CommandMetadata {
	return map[string]CommandMetadata{
		"compact": {Category: "context", ArgumentHint: "[now|auto|off|status]", Keywords: []string{"context", "trim", "token"}, Examples: []string{"/compact now", "/compact status"}},
		"context": {Category: "context", ArgumentHint: "[show|clear]", Keywords: []string{"flags", "state"}, Examples: []string{"/context show"}},
		"memory":  {Category: "context", ArgumentHint: "[list|add|remove|clear|status]", Keywords: []string{"notes", "scratchpad"}, Examples: []string{"/memory add api requires auth"}},
		"files":   {Category: "context", ArgumentHint: "[list|add|remove|clear|status]", Keywords: []string{"attachments", "paths"}, Examples: []string{"/files add README.md"}},
		"env":     {Category: "context", ArgumentHint: "[list|set|unset|status]", Keywords: []string{"environment", "vars"}, Examples: []string{"/env set FOO bar"}},
		"advisor": {Category: "context", ArgumentHint: "[status|set <model>]", Keywords: []string{"advisor", "model"}, Examples: []string{"/advisor set openai/gpt-4o"}},
		"btw":     {Category: "context", ArgumentHint: "[status|ask <question>]", Keywords: []string{"question", "notes"}, Examples: []string{"/btw ask what changed"}},
	}
}
