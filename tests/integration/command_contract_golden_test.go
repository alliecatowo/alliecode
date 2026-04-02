package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/permissions"
)

func TestCommandContractGoldenOutputs(t *testing.T) {
	t.Parallel()

	registry := commands.DefaultRegistry()
	state := &commands.RuntimeState{
		CostInputTokens:  11,
		CostOutputTokens: 22,
		CostCacheRead:    3,
		CostCacheWrite:   4,
		OutputStyle:      "dark",
		CompactMode:      "auto",
		CompactCount:     5,
		ResumeCount:      6,
		CopyCount:        7,
		ClearCount:       8,
		BranchCount:      9,
		DiffCount:        10,
		MemoryEntries:    []string{"remember alpha", "remember beta"},
		MemoryWrites:     11,
		MemoryClears:     2,
		LastMemory:       "remember beta",
		PrivacyTelemetry: true,
		PrivacyTraining:  false,
		HookPreEnabled:   true,
		HookPostEnabled:  false,
		SandboxMode:      "read-only",
		Tasks:            []string{"task one", "task two"},
		TasksCompleted:   12,
		ConfigValues: map[string]string{
			"z": "last",
			"a": "first",
		},
		PermissionMode:  permissions.ModeDefault,
		PermissionRules: []string{"write", "read"},
		HistoryEntries: []commands.HistoryEntry{
			{ID: "b", CreatedAt: "2026-03-01T00:00:00Z", Model: "ollama/llama3", Turns: 2, Title: "second", Summary: "follow-up"},
			{ID: "a", CreatedAt: "2026-02-01T00:00:00Z", Model: "openai/gpt-4o-mini", Turns: 1, Title: " first title ", Summary: " session summary "},
		},
		ProviderName:  "ollama",
		ProviderReady: false,
		ConfigPath:    "/tmp/config.yml",
		MCPConnections: map[string]bool{
			"zeta":  false,
			"alpha": true,
		},
		ReviewCount:       13,
		LastReviewTarget:  "org/repo#456",
		CommitCount:       14,
		LastCommitMessage: "test: runtime parity coverage",
		CommitPushPRCount: 15,
		LastPRURL:         "https://example.test/pr/42",
	}

	tests := []struct {
		name       string
		input      string
		goldenFile string
	}{
		{name: "cost", input: "/cost", goldenFile: "cost.golden"},
		{name: "vim_status", input: "/vim status", goldenFile: "vim_status.golden"},
		{name: "voice_status", input: "/voice status", goldenFile: "voice_status.golden"},
		{name: "theme", input: "/theme", goldenFile: "theme.golden"},
		{name: "status", input: "/status", goldenFile: "status.golden"},
		{name: "stats", input: "/stats", goldenFile: "stats.golden"},
		{name: "memory", input: "/memory", goldenFile: "memory.golden"},
		{name: "privacy_settings", input: "/privacy-settings", goldenFile: "privacy_settings.golden"},
		{name: "hooks", input: "/hooks", goldenFile: "hooks.golden"},
		{name: "sandbox", input: "/sandbox", goldenFile: "sandbox.golden"},
		{name: "sandbox_check", input: "/sandbox check", goldenFile: "sandbox_check.golden"},
		{name: "tasks", input: "/tasks", goldenFile: "tasks.golden"},
		{name: "config_show", input: "/config show", goldenFile: "config_show.golden"},
		{name: "config_get", input: "/config get a", goldenFile: "config_get.golden"},
		{name: "usage", input: "/usage", goldenFile: "usage.golden"},
		{name: "version", input: "/version", goldenFile: "version.golden"},
		{name: "context_show", input: "/context show", goldenFile: "context_show.golden"},
		{name: "agents_list", input: "/agents list", goldenFile: "agents_list.golden"},
		{name: "clear_status", input: "/clear status", goldenFile: "clear_status.golden"},
		{name: "permissions_status", input: "/permissions status", goldenFile: "permissions_status.golden"},
		{name: "permissions_rules", input: "/permissions rules", goldenFile: "permissions_rules.golden"},
		{name: "history_list", input: "/history list", goldenFile: "history_list.golden"},
		{name: "history_show", input: "/history show a", goldenFile: "history_show.golden"},
		{name: "doctor_human", input: "/doctor human", goldenFile: "doctor_human.golden"},
		{name: "doctor_json", input: "/doctor json", goldenFile: "doctor_json.golden"},
		{name: "commit_status", input: "/commit status", goldenFile: "commit_status.golden"},
		{name: "commit_push_pr_status", input: "/commit-push-pr status", goldenFile: "commit_push_pr_status.golden"},
		{name: "mcp_list", input: "/mcp list", goldenFile: "mcp_list.golden"},
		{name: "mcp_status", input: "/mcp status alpha", goldenFile: "mcp_status.golden"},
		{name: "mcp_status_disconnected", input: "/mcp status zeta", goldenFile: "mcp_status_disconnected.golden"},
		{name: "add_dir", input: "/add-dir /", goldenFile: "add_dir.golden"},
		{name: "model", input: "/model", goldenFile: "model.golden"},
		{name: "compact_status", input: "/compact status", goldenFile: "compact_status.golden"},
		{name: "resume_status", input: "/resume status", goldenFile: "resume_status.golden"},
		{name: "branch_status", input: "/branch status", goldenFile: "branch_status.golden"},
		{name: "branch_list", input: "/branch list", goldenFile: "branch_list.golden"},
		{name: "diff_status", input: "/diff status", goldenFile: "diff_status.golden"},
		{name: "diff_list", input: "/diff", goldenFile: "diff_list.golden"},
		{name: "init", input: "/init", goldenFile: "init.golden"},
		{name: "review_status", input: "/review status", goldenFile: "review_status.golden"},
		{name: "files_status", input: "/files status", goldenFile: "files_status.golden"},
		{name: "files_list", input: "/files list", goldenFile: "files_list.golden"},
		{name: "output_style_status", input: "/output-style status", goldenFile: "output_style_status.golden"},
		{name: "output_style_list", input: "/output-style list", goldenFile: "output_style_list.golden"},
		{name: "statusline_status", input: "/statusline status", goldenFile: "statusline_status.golden"},
		{name: "keybindings_status", input: "/keybindings status", goldenFile: "keybindings_status.golden"},
		{name: "keybindings_path", input: "/keybindings path", goldenFile: "keybindings_path.golden"},
		{name: "upgrade_status", input: "/upgrade status", goldenFile: "upgrade_status.golden"},
		{name: "terminal_setup_status", input: "/terminal-setup status", goldenFile: "terminal_setup_status.golden"},
		{name: "terminal_setup_detect", input: "/terminal-setup detect", goldenFile: "terminal_setup_detect.golden"},
		{name: "terminal_setup_apply", input: "/terminal-setup apply zed", goldenFile: "terminal_setup_apply.golden"},
		{name: "release_notes_status", input: "/release-notes status", goldenFile: "release_notes_status.golden"},
		{name: "install_github_app_status", input: "/install-github-app status", goldenFile: "install_github_app_status.golden"},
		{name: "install_slack_app_status", input: "/install-slack-app status", goldenFile: "install_slack_app_status.golden"},
		{name: "feedback_status", input: "/feedback status", goldenFile: "feedback_status.golden"},
		{name: "advisor_status", input: "/advisor status", goldenFile: "advisor_status.golden"},
		{name: "login_status", input: "/login status", goldenFile: "login_status.golden"},
		{name: "logout_status", input: "/logout status", goldenFile: "logout_status.golden"},
		{name: "chrome_status", input: "/chrome status", goldenFile: "chrome_status.golden"},
		{name: "color_status", input: "/color", goldenFile: "color_status.golden"},
		{name: "desktop_status", input: "/desktop status", goldenFile: "desktop_status.golden"},
		{name: "mobile_status", input: "/mobile status", goldenFile: "mobile_status.golden"},
		{name: "fast_status", input: "/fast status", goldenFile: "fast_status.golden"},
		{name: "effort_status", input: "/effort status", goldenFile: "effort_status.golden"},
		{name: "plugin_status", input: "/plugin status", goldenFile: "plugin_status.golden"},
		{name: "reload_plugins", input: "/reload-plugins", goldenFile: "reload_plugins.golden"},
		{name: "export", input: "/export", goldenFile: "export.golden"},
		{name: "extra_usage_status", input: "/extra-usage status", goldenFile: "extra_usage_status.golden"},
		{name: "rate_limit_options_status", input: "/rate-limit-options status", goldenFile: "rate_limit_options_status.golden"},
		{name: "pr_comments_status", input: "/pr-comments status", goldenFile: "pr_comments_status.golden"},
		{name: "web_setup_status", input: "/web-setup status", goldenFile: "web_setup_status.golden"},
		{name: "btw_result", input: "/btw quick check", goldenFile: "btw_result.golden"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := registry.Dispatch(context.Background(), commands.Context{State: state}, tc.input)
			if err != nil {
				t.Fatalf("dispatch %s error = %v", tc.input, err)
			}
			if !res.Handled {
				t.Fatalf("dispatch %s returned Handled=false", tc.input)
			}
			want := normalizeContractOutput(tc.name, readGolden(t, tc.goldenFile))
			got := normalizeContractOutput(tc.name, res.Message)
			if got != want {
				t.Fatalf("contract mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", tc.input, got, want)
			}
		})
	}
}

func normalizeContractOutput(name, msg string) string {
	if name != "commit_status" {
		return msg
	}

	lines := strings.Split(msg, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, "files=") ||
			strings.HasPrefix(line, "staged=") ||
			strings.HasPrefix(line, "unstaged=") ||
			strings.HasPrefix(line, "untracked=") {
			continue
		}
		filtered = append(filtered, line)
	}

	return strings.Join(filtered, "\n")
}

func readGolden(t *testing.T, name string) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("runtime.Caller(0) failed")
	}
	path := filepath.Join(filepath.Dir(filename), "testdata", "command_contract", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", path, err)
	}
	return strings.TrimSuffix(string(b), "\n")
}
