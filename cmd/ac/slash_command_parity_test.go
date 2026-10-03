package main

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func TestRunSlashCommandExpandedParitySurface(t *testing.T) {
	cmd := &cobra.Command{}
	out := &strings.Builder{}
	cmd.SetOut(out)
	state := commands.RuntimeState{}

	cases := []struct {
		input string
		want  string
	}{
		{input: "/bridge status", want: "BRIDGE_STATUS"},
		{input: "/autofix-pr status", want: "AUTOFIX_PR_STATUS"},
		{input: "/backfill-sessions status", want: "BACKFILL_STATUS"},
		{input: "/break-cache status", want: "BREAK_CACHE_STATUS"},
		{input: "/bughunter status", want: "BUGHUNTER_STATUS"},
		{input: "/ctx-viz status", want: "CTX_VIZ_STATUS"},
		{input: "/debug-tool-call status", want: "DEBUG_TOOL_CALL_STATUS"},
		{input: "/good-claude status", want: "GOOD_CLAUDE_STATUS"},
		{input: "/heapdump status", want: "HEAPDUMP_STATUS"},
		{input: "/install status", want: "INSTALL_STATUS"},
		{input: "/mock-limits status", want: "MOCK_LIMITS_STATUS"},
		{input: "/onboarding status", want: "ONBOARDING_STATUS"},
		{input: "/perf-issue status", want: "PERF_ISSUE_STATUS"},
		{input: "/sandbox-toggle status", want: "SANDBOX_TOGGLE_STATUS"},
		{input: "/remote-setup status", want: "REMOTE_SETUP_STATUS"},
		{input: "/ultraplan status", want: "ULTRAPLAN_STATUS"},
	}

	for _, tc := range cases {
		out.Reset()
		if err := runSlashCommand(cmd.Context(), cmd, tc.input, state); err != nil {
			t.Fatalf("runSlashCommand(%q) error = %v", tc.input, err)
		}
		if !strings.Contains(out.String(), tc.want) {
			t.Fatalf("expected %q output for %q, got %q", tc.want, tc.input, out.String())
		}
	}
}
