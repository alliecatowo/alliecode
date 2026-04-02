package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func TestCommandContractWorkflowOutputs(t *testing.T) {
	t.Parallel()

	registry := commands.DefaultRegistry()

	tests := []struct {
		name       string
		input      string
		goldenFile string
	}{
		{name: "plan_status", input: "/plan status", goldenFile: "workflow_plan.golden"},
		{name: "session_status", input: "/session status", goldenFile: "workflow_session.golden"},
		{name: "skills_status", input: "/skills status", goldenFile: "workflow_skills.golden"},
		{name: "rewind_status", input: "/rewind status", goldenFile: "workflow_rewind.golden"},
		{name: "tag_status", input: "/tag status", goldenFile: "workflow_tag.golden"},
		{name: "remote_env_status", input: "/remote-env status", goldenFile: "workflow_remote_env.golden"},
		{name: "security_review_status", input: "/security-review status", goldenFile: "workflow_security_review.golden"},
		{name: "exit", input: "/exit", goldenFile: "workflow_exit_status.golden"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			res, err := registry.Dispatch(context.Background(), commands.Context{State: workflowState()}, tc.input)
			if err != nil {
				t.Fatalf("dispatch %s error = %v", tc.input, err)
			}
			want := readWorkflowGolden(t, tc.goldenFile)
			if res.Message != want {
				t.Fatalf("contract mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", tc.input, res.Message, want)
			}
		})
	}
}

func workflowState() *commands.RuntimeState {
	return &commands.RuntimeState{
		PlanModeEnabled:     true,
		PlanEnableCount:     2,
		PlanOpenCount:       3,
		LastPlanDescription: "draft p1",
		TransportMode:       "remote",
		SessionID:           "session-startup-42",
		SessionPath:         "/tmp/.alliecode/sessions/session-startup-42.jsonl",
		RemoteSessionURL:    "https://example.test/session/abc",
		SessionViewCount:    4,
		Skills:              []string{"refactor", "debug"},
		SkillsViewCount:     5,
		RewindCount:         6,
		LastRewindTarget:    "checkpoint-1",
		CurrentTag:          "release",
		TagUpdates:          7,
		RemoteEnvironment:   "hardened-linux",
		RemoteEnvUpdates:    8,
		SecurityReviewCount: 9,
		LastSecurityTarget:  "feature/auth",
		ExitCount:           10,
	}
}

func readWorkflowGolden(t *testing.T, name string) string {
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
