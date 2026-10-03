package commands

import (
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

type PanelApplyMode string

const (
	PanelApplySubmit PanelApplyMode = "submit"
	PanelApplyStage  PanelApplyMode = "stage"
)

type InteractivePanel struct {
	Command       string
	Title         string
	Subtitle      string
	HeaderIntents []types.RenderIntent
	FooterIntents []types.RenderIntent
	Items         []InteractivePanelItem
}

type InteractivePanelItem struct {
	Key            string
	Section        string
	Label          string
	Detail         string
	Status         string
	Preview        []string
	PreviewIntents []types.RenderIntent
	ApplyInput     string
	ApplyMode      PanelApplyMode
}

func SupportsInteractivePanel(name string) bool {
	switch canonicalInteractivePanelName(name) {
	case "model", "permissions", "doctor", "status", "mcp", "plugin", "skills", "history", "session":
		return true
	case "provider":
		return true
	case "branch", "diff", "files", "memory", "theme", "output-style", "privacy-settings", "upgrade", "resume", "plan":
		return true
	case "bridge", "sandbox-toggle", "mock-limits", "onboarding", "remote-setup":
		return true
	case "tasks", "env", "issue", "workflows", "proactive", "assistant", "share", "compact":
		return true
	default:
		return false
	}
}

func InteractivePanelForCommand(name string, cmdCtx Context) (InteractivePanel, bool) {
	switch canonicalInteractivePanelName(name) {
	case "model":
		return buildModelInteractivePanel(cmdCtx), true
	case "provider":
		return buildProviderInteractivePanel(cmdCtx), true
	case "permissions":
		return buildPermissionsInteractivePanel(cmdCtx), true
	case "doctor":
		return buildDoctorInteractivePanel(cmdCtx), true
	case "status":
		return buildStatusInteractivePanel(cmdCtx), true
	case "mcp":
		return buildMCPInteractivePanel(cmdCtx), true
	case "plugin":
		return buildPluginInteractivePanel(cmdCtx), true
	case "skills":
		return buildSkillsInteractivePanel(cmdCtx), true
	case "history":
		return buildHistoryInteractivePanel(cmdCtx), true
	case "session":
		return buildSessionInteractivePanel(cmdCtx), true
	case "branch":
		return buildBranchInteractivePanel(cmdCtx), true
	case "diff":
		return buildDiffInteractivePanel(cmdCtx), true
	case "files":
		return buildFilesInteractivePanel(cmdCtx), true
	case "memory":
		return buildMemoryInteractivePanel(cmdCtx), true
	case "theme":
		return buildThemeInteractivePanel(cmdCtx), true
	case "output-style":
		return buildOutputStyleInteractivePanel(cmdCtx), true
	case "privacy-settings":
		return buildPrivacyInteractivePanel(cmdCtx), true
	case "upgrade":
		return buildUpgradeInteractivePanel(cmdCtx), true
	case "resume":
		return buildResumeInteractivePanel(cmdCtx), true
	case "plan":
		return buildPlanInteractivePanel(cmdCtx), true
	case "bridge":
		return buildBridgeInteractivePanel(cmdCtx), true
	case "sandbox-toggle":
		return buildSandboxToggleInteractivePanel(cmdCtx), true
	case "mock-limits":
		return buildMockLimitsInteractivePanel(cmdCtx), true
	case "onboarding":
		return buildOnboardingInteractivePanel(cmdCtx), true
	case "remote-setup":
		return buildRemoteSetupInteractivePanel(cmdCtx), true
	case "tasks":
		return buildTasksInteractivePanel(cmdCtx), true
	case "env":
		return buildEnvInteractivePanel(cmdCtx), true
	case "issue":
		return buildIssueInteractivePanel(cmdCtx), true
	case "workflows":
		return buildWorkflowsInteractivePanel(cmdCtx), true
	case "proactive":
		return buildProactiveInteractivePanel(cmdCtx), true
	case "assistant":
		return buildAssistantInteractivePanel(cmdCtx), true
	case "share":
		return buildShareInteractivePanel(cmdCtx), true
	case "compact":
		return buildCompactInteractivePanel(cmdCtx), true
	default:
		return InteractivePanel{}, false
	}
}

func canonicalInteractivePanelName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func interactivePanelState(cmdCtx Context) *RuntimeState {
	if cmdCtx.State == nil {
		return &RuntimeState{}
	}
	return cmdCtx.State
}

func previewLines(lines ...string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
	}
	return out
}
