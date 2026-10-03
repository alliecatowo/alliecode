package commands

func gitMetadata() map[string]CommandMetadata {
	return map[string]CommandMetadata{
		"branch":         {Category: "git", Group: "workspace", ArgumentHint: "[status|list|create|switch]", HelpHint: "Switch after create to update active branch", Shortcuts: []string{"br"}, Keywords: []string{"git", "checkout"}, Examples: []string{"/branch list", "/branch create feature/x"}},
		"diff":           {Category: "git", Group: "workspace", ArgumentHint: "[status|list|add|clear|mode]", HelpHint: "Use /diff list before /commit", Shortcuts: []string{"df"}, Keywords: []string{"changes", "patch"}, Examples: []string{"/diff status"}},
		"commit":         {Category: "git", Group: "delivery", ArgumentHint: "[status|suggest|check|<message>]", HelpHint: "Use /commit suggest for draft message", Shortcuts: []string{"ci"}, Keywords: []string{"message", "staging"}, Examples: []string{"/commit", "/commit suggest", "/commit fix parser edge case"}},
		"commit-push-pr": {Category: "git", Group: "delivery", ArgumentHint: "[status|doctor|--force] [--base <branch>] [title]", HelpHint: "Run /commit-push-pr doctor before --force", Shortcuts: []string{"cpr"}, Keywords: []string{"pull request", "push"}, Examples: []string{"/commit-push-pr status", "/commit-push-pr doctor", "/commit-push-pr --base main"}},
		"review":         {Category: "git", Group: "review", ArgumentHint: "[status [target]|focus <risk|perf|tests|security>|<target>]", HelpHint: "Use /review focus security for threat pass", Shortcuts: []string{"rv"}, Keywords: []string{"pr", "code review"}, Examples: []string{"/review #42", "/review focus tests"}},
		"pr-comments":    {Category: "git", Group: "review", ArgumentHint: "[status|<ref>|file=<path>|state]", HelpHint: "Use file=<path> to replay local review threads", Shortcuts: []string{"comments"}, Keywords: []string{"review comments", "threads"}, Examples: []string{"/pr-comments status"}},
		"autofix-pr":     {Category: "git", Group: "delivery", ArgumentHint: "[status|run|plan|cancel]", HelpHint: "Use plan before run for risky diffs", Keywords: []string{"autofix", "pull request"}, Examples: []string{"/autofix-pr plan #42"}},
	}
}

func workflowMetadata() map[string]CommandMetadata {
	return map[string]CommandMetadata{
		"issue":             {Category: "workflow", Group: "tracker", ArgumentHint: "[status|list|provider|create|open|close|assign|label|unlabel]", HelpHint: "Use labels to drive workflow filters", Shortcuts: []string{"is"}, Keywords: []string{"ticket", "bug"}, Examples: []string{"/issue create fix flaky test", "/issue label ISSUE-1 bug"}},
		"workflows":         {Category: "workflow", Group: "tracker", ArgumentHint: "[status|list|run|complete|fail|cancel|rerun]", HelpHint: "Use rerun after fail to keep history coherent", Shortcuts: []string{"wf"}, Keywords: []string{"ci", "automation"}, Examples: []string{"/workflows run ci", "/workflows rerun ci"}},
		"tasks":             {Category: "workflow", Group: "tracker", ArgumentHint: "[list|add|done|clear]", HelpHint: "Track short milestones tied to commands", Shortcuts: []string{"todo"}, Keywords: []string{"todo", "tracking"}, Examples: []string{"/tasks add ship release notes"}},
		"assistant":         {Category: "workflow", Group: "runtime", ArgumentHint: "[status|mode|session|reset]", HelpHint: "Use review mode before /review passes", Shortcuts: []string{"assist"}, Keywords: []string{"modes", "chat"}, Examples: []string{"/assistant mode review"}},
		"proactive":         {Category: "workflow", Group: "runtime", ArgumentHint: "[status|on|off|rule|trigger]", HelpHint: "Add narrow rules before enabling globally", Shortcuts: []string{"pro"}, Keywords: []string{"automation", "rules"}, Examples: []string{"/proactive on"}},
		"backfill-sessions": {Category: "workflow", Group: "runtime", ArgumentHint: "[status|run|dry-run]", HelpHint: "Backfill historical sessions deterministically", Keywords: []string{"history", "migrate"}, Examples: []string{"/backfill-sessions run 20"}},
		"perf-issue":        {Category: "workflow", Group: "tracker", ArgumentHint: "[status|open <title>]", HelpHint: "Capture performance regressions with titles", Keywords: []string{"performance", "issue"}, Examples: []string{"/perf-issue open slow startup"}},
		"ultraplan":         {Category: "workflow", Group: "runtime", ArgumentHint: "[status|run [target]|clear]", HelpHint: "Use for deep planning workflows", Keywords: []string{"planning", "strategy"}, Examples: []string{"/ultraplan run release"}},
	}
}
