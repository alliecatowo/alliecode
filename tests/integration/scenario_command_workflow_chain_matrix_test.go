package integration_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/permissions"
)

type commandWorkflowCase struct {
	Name         string `json:"name"`
	Input        string `json:"input"`
	StateProfile string `json:"state_profile"`
	GoldenPath   string `json:"golden_path"`
	NetworkOnly  bool   `json:"network_only"`
}

func TestScenarioMatrix_CommandWorkflowChains(t *testing.T) {
	t.Parallel()

	registry := commands.DefaultRegistry()
	paths := loadScenarioMatrixPaths(t, "command_workflow_chain", "*.json")
	for _, path := range paths {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()

			var tc commandWorkflowCase
			decodeScenarioCase(t, path, &tc)
			if tc.NetworkOnly {
				t.Skip("network-only scenario")
			}

			state := scenarioCommandState(tc.StateProfile)
			res, err := registry.Dispatch(context.Background(), commands.Context{State: state}, tc.Input)
			if err != nil {
				t.Fatalf("dispatch %s error = %v", tc.Input, err)
			}
			if !res.Handled {
				t.Fatalf("dispatch %s returned Handled=false", tc.Input)
			}

			want := normalizeContractOutput(tc.Name, readWorkflowGolden(t, tc.GoldenPath))
			got := normalizeContractOutput(tc.Name, res.Message)
			if got != want {
				t.Fatalf("contract mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", tc.Input, got, want)
			}
		})
	}
}

func scenarioCommandState(profile string) *commands.RuntimeState {
	switch profile {
	case "workflow":
		return workflowState()
	case "contract":
		return &commands.RuntimeState{
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
	default:
		return &commands.RuntimeState{}
	}
}
