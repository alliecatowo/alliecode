package commands

func generalMetadata() map[string]CommandMetadata {
	return map[string]CommandMetadata{
		"help":    {Category: "general", Group: "core", ArgumentHint: "[query]", HelpHint: "Type part of a command name or keyword", Shortcuts: []string{"?"}, Keywords: []string{"discover", "docs", "usage"}, Examples: []string{"/help", "/help model"}},
		"version": {Category: "general", Group: "core", ArgumentHint: "", HelpHint: "Useful for bug reports", Shortcuts: []string{"ver"}, Keywords: []string{"build", "version"}, Examples: []string{"/version"}},
		"exit":    {Category: "general", Group: "core", ArgumentHint: "", HelpHint: "Same as /quit alias", Shortcuts: []string{"quit"}, Keywords: []string{"quit", "close"}, Examples: []string{"/exit"}},
	}
}

func authMetadata() map[string]CommandMetadata {
	return map[string]CommandMetadata{
		"login":         {Category: "auth", Group: "identity", ArgumentHint: "[status|provider|account]", HelpHint: "Provider can be passed directly: /login openai", Shortcuts: []string{"signin"}, Keywords: []string{"signin", "oauth"}, Examples: []string{"/login status", "/login provider openai"}},
		"logout":        {Category: "auth", Group: "identity", ArgumentHint: "[status]", HelpHint: "Clears in-memory login state", Shortcuts: []string{"signout"}, Keywords: []string{"signout", "auth"}, Examples: []string{"/logout"}},
		"oauth-refresh": {Category: "auth", Group: "identity", ArgumentHint: "[status|refresh|error|clear]", HelpHint: "Useful when provider sessions expire", Shortcuts: []string{"oauth"}, Keywords: []string{"tokens", "oauth"}, Examples: []string{"/oauth-refresh refresh github"}},
	}
}

func accountMetadata() map[string]CommandMetadata {
	return map[string]CommandMetadata{
		"upgrade":            {Category: "account", ArgumentHint: "[status|max|pro|team|enterprise]", Keywords: []string{"plan", "billing"}, Examples: []string{"/upgrade pro"}},
		"extra-usage":        {Category: "account", ArgumentHint: "[status|enable|disable|request]", Keywords: []string{"limits", "quota"}, Examples: []string{"/extra-usage request"}},
		"rate-limit-options": {Category: "account", ArgumentHint: "[status|upgrade|extra-usage|cancel]", Keywords: []string{"limits", "quota"}, Examples: []string{"/rate-limit-options status"}},
		"usage":              {Category: "account", ArgumentHint: "", Keywords: []string{"snapshot", "usage"}, Examples: []string{"/usage"}},
		"cost":               {Category: "account", ArgumentHint: "", Keywords: []string{"tokens", "cost"}, Examples: []string{"/cost"}},
	}
}
