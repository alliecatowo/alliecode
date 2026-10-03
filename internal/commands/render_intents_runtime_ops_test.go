package commands

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/permissions"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestRuntimeOpsCommandsEmitRenderIntents(t *testing.T) {
	baseState := &RuntimeState{
		PermissionMode: permissions.ModeAuto,
		Branches:       []string{"main", "feature/ui"},
		ActiveBranch:   "main",
		DiffEntries:    []DiffEntry{{Path: "internal/tui/app.go", Added: 10, Removed: 2}},
		ConfigValues: map[string]string{
			"settings.output-style":  "default",
			"settings.output-format": "default",
			"settings.transport":     "local",
		},
		ContextFiles:     []string{"internal/tui/app.go"},
		MemoryEntries:    []string{"use typed intents"},
		PrivacyTelemetry: true,
		OutputStyle:      "system",
		OutputFormat:     "default",
	}

	tests := []struct {
		name string
		cmd  Command
		ctx  Context
		inv  Invocation
		kind types.RenderIntentKind
	}{
		{name: "resume status", cmd: NewResumeCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "resume", Args: []string{"status"}}, kind: types.RenderIntentSummaryCard},
		{name: "branch list", cmd: NewBranchCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "branch", Args: []string{"list"}}, kind: types.RenderIntentTable},
		{name: "diff status", cmd: NewDiffCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "diff", Args: []string{"status"}}, kind: types.RenderIntentSummaryCard},
		{name: "cost", cmd: NewCostCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "cost"}, kind: types.RenderIntentSummaryCard},
		{name: "usage", cmd: NewUsageCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "usage"}, kind: types.RenderIntentSummaryCard},
		{name: "context show", cmd: NewContextCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "context", Args: []string{"show"}}, kind: types.RenderIntentDetailRows},
		{name: "clear status", cmd: NewClearCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "clear", Args: []string{"status"}}, kind: types.RenderIntentSummaryCard},
		{name: "files list", cmd: NewFilesCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "files", Args: []string{"list"}}, kind: types.RenderIntentTable},
		{name: "theme list", cmd: NewThemeCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "theme", Args: []string{"list"}}, kind: types.RenderIntentOptionList},
		{name: "output style list", cmd: NewOutputStyleCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "output-style", Args: []string{"list"}}, kind: types.RenderIntentOptionList},
		{name: "memory list", cmd: NewMemoryCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "memory", Args: []string{"list"}}, kind: types.RenderIntentChecklist},
		{name: "privacy list", cmd: NewPrivacySettingsCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "privacy-settings", Args: []string{"list"}}, kind: types.RenderIntentOptionList},
		{name: "upgrade status", cmd: NewUpgradeCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "upgrade", Args: []string{"status"}}, kind: types.RenderIntentSummaryCard},
		{name: "plan status", cmd: NewPlanCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "plan", Args: []string{"status"}}, kind: types.RenderIntentSummaryCard},
		{name: "copy", cmd: NewCopyCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "copy", Args: []string{"hello", "world"}}, kind: types.RenderIntentDetailRows},
		{name: "version", cmd: NewVersionCommand(), ctx: Context{}, inv: Invocation{Name: "version"}, kind: types.RenderIntentSummaryCard},
		{name: "rewind status", cmd: NewRewindCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "rewind", Args: []string{"status"}}, kind: types.RenderIntentSummaryCard},
		{name: "tag status", cmd: NewTagCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "tag", Args: []string{"status"}}, kind: types.RenderIntentSummaryCard},
		{name: "remote env list", cmd: NewRemoteEnvCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "remote-env", Args: []string{"list"}}, kind: types.RenderIntentOptionList},
		{name: "security review status", cmd: NewSecurityReviewCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "security-review", Args: []string{"status"}}, kind: types.RenderIntentSummaryCard},
		{name: "agents list", cmd: NewAgentsCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "agents", Args: []string{"list"}}, kind: types.RenderIntentTable},
		{name: "vim status", cmd: NewVimCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "vim", Args: []string{"status"}}, kind: types.RenderIntentSummaryCard},
		{name: "voice status", cmd: NewVoiceCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "voice", Args: []string{"status"}}, kind: types.RenderIntentSummaryCard},
		{name: "buddy help", cmd: NewBuddyCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "buddy", Args: []string{"help"}}, kind: types.RenderIntentOptionList},
		{name: "statusline status", cmd: NewStatuslineCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "statusline", Args: []string{"status"}}, kind: types.RenderIntentSummaryCard},
		{name: "ide detect", cmd: NewIDECommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "ide", Args: []string{"detect"}}, kind: types.RenderIntentOptionList},
		{name: "keybindings status", cmd: NewKeybindingsCommand(), ctx: Context{State: cloneState(baseState)}, inv: Invocation{Name: "keybindings", Args: []string{"status"}}, kind: types.RenderIntentSummaryCard},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tc.cmd.Execute(context.Background(), tc.ctx, tc.inv)
			if err != nil {
				t.Fatalf("execute failed: %v", err)
			}
			if len(res.RenderIntents) == 0 {
				t.Fatalf("expected render intents, got none (message=%q)", res.Message)
			}
			found := false
			for _, intent := range res.RenderIntents {
				if intent.Kind == tc.kind {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("expected intent kind %q in %#v", tc.kind, res.RenderIntents)
			}
		})
	}
}

func cloneState(in *RuntimeState) *RuntimeState {
	if in == nil {
		return nil
	}
	out := *in
	out.Branches = append([]string(nil), in.Branches...)
	out.DiffEntries = append([]DiffEntry(nil), in.DiffEntries...)
	out.ContextFiles = append([]string(nil), in.ContextFiles...)
	out.MemoryEntries = append([]string(nil), in.MemoryEntries...)
	out.ConfigValues = map[string]string{}
	for k, v := range in.ConfigValues {
		out.ConfigValues[k] = v
	}
	return &out
}
