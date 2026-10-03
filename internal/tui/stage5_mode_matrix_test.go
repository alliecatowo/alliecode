package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func TestStage5ModeMatrixPermissionPrecedesAllInteractiveModes(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.state = stateSearch
	app.searchMode = searchModeQuickOpen
	app.commandPanel.active = true
	app.slashAutocomplete.items = []commands.Suggestion{{Name: "status"}}
	app.refAuto.active = true
	app.permissionQueue = []permissionPromptRequest{{toolName: "bash", description: "run"}}
	app.ensurePermissionPromptVisible()
	if app.inputMode != inputModePermission {
		t.Fatalf("expected permission mode precedence, got %s", app.inputMode)
	}
}

func TestStage5ModeMatrixQuickOpenHistoryTimelineContexts(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app.startQuickOpen()
	if got := app.activeContextHint(); !strings.HasPrefix(got, "search ") {
		t.Fatalf("expected quick-open context, got %q", got)
	}
	app.startHistorySearch()
	if got := app.activeContextHint(); !strings.HasPrefix(got, "search ") {
		t.Fatalf("expected history context, got %q", got)
	}
	app.startTimelineSearch()
	if got := app.activeContextHint(); !strings.HasPrefix(got, "search ") {
		t.Fatalf("expected timeline context, got %q", got)
	}
}
