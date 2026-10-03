package e2e_test

import (
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/tui"
)

func TestE2ECommandPanelFlowProviderOneEnterOpenThenApply(t *testing.T) {
	app := e2eReadyApp(t, 180, 30)
	app = e2eTypeText(t, app, "/provider")
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter})

	openView := e2ePlainView(app)
	if !strings.Contains(openView, "drawer: provider panel: /provider") {
		t.Fatalf("expected /provider one-enter to open command panel")
	}

	app = e2eSendKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEnter})
	afterView := e2ePlainView(app)
	if strings.Contains(afterView, "drawer: provider panel: /provider") {
		t.Fatalf("expected panel to close after apply")
	}
	if !strings.Contains(afterView, "Current provider availability for this session.") {
		t.Fatalf("expected apply to execute provider command, got:\n%s", afterView)
	}
}

func TestE2ECommandPanelFlowTasksOpenWithoutAccidentalRun(t *testing.T) {
	app := e2eReadyApp(t, 180, 30)
	app = e2eTypeText(t, app, "/tasks")
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter})

	view := e2ePlainView(app)
	if !strings.Contains(view, "drawer: tasks panel: /tasks") {
		t.Fatalf("expected /tasks one-enter to open command panel")
	}
	if strings.Contains(view, "TASKS_LIST") {
		t.Fatalf("expected no accidental run when panel opens, got:\n%s", view)
	}
}

func TestE2ECommandPanelFlowArgCommandStagesAndDoesNotRun(t *testing.T) {
	app := e2eReadyApp(t, 180, 30)
	app = e2eTypeText(t, app, "/model")
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter})

	view := e2ePlainView(app)
	if !strings.Contains(view, "model picker: /model") {
		t.Fatalf("expected /model one-enter to stage and open picker")
	}
	if strings.Contains(view, "drawer: model panel: /model") {
		t.Fatalf("expected model flow to stage args, not open panel")
	}
	if strings.Contains(view, "Model set to") || strings.Contains(view, "MODEL_STATUS") {
		t.Fatalf("expected staged arg command to avoid accidental run, got:\n%s", view)
	}
}

func TestE2ECommandPanelFlowIssueOpenThenEscapeThenReopen(t *testing.T) {
	app := e2eReadyApp(t, 180, 30)
	app = e2eTypeText(t, app, "/issue")
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter})

	openView := e2ePlainView(app)
	if !strings.Contains(openView, "drawer: issue panel: /issue") {
		t.Fatalf("expected /issue one-enter to open issue panel")
	}
	if strings.Contains(openView, "ISSUE_LIST") || strings.Contains(openView, "ISSUE_STATUS") {
		t.Fatalf("expected no legacy issue dump while panel is open, got:\n%s", openView)
	}

	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyEsc})
	escView := e2ePlainView(app)
	if strings.Contains(escView, "drawer: issue panel: /issue") {
		t.Fatalf("expected escape to close issue panel")
	}
	if !strings.Contains(escView, "input mode: chat") {
		t.Fatalf("expected chat mode after escaping issue panel")
	}

}

func TestE2ECommandPanelFlowPermissionsOpenThenTabShiftTabStillPanelized(t *testing.T) {
	app := e2eReadyApp(t, 180, 30)
	app = e2eTypeText(t, app, "/perm")
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter})
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyTab})
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab})

	view := e2ePlainView(app)
	if !strings.Contains(view, "input mode: command-panel") || !strings.Contains(view, "command panel /permissions") {
		t.Fatalf("expected permissions panel context after tab cycle, got:\n%s", view)
	}
	if strings.Contains(view, "PERMISSIONS_MODE") || strings.Contains(view, "PERMISSIONS_MODES") {
		t.Fatalf("expected no legacy permissions contract while panelized, got:\n%s", view)
	}
}

func TestE2ECommandPanelFlowStatusThenTasksPanelChain(t *testing.T) {
	app := e2eReadyApp(t, 180, 30)
	app = e2eTypeText(t, app, "/status")
	app = e2eSendKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEnter})

	statusView := e2ePlainView(app)
	if strings.Contains(statusView, "STATUS_REPORT") {
		t.Fatalf("expected /status to render intent output, got:\n%s", statusView)
	}
	if !strings.Contains(statusView, "Provider:") || !strings.Contains(statusView, "Tasks:") {
		t.Fatalf("expected status intent content, got:\n%s", statusView)
	}

	app = e2eTypeText(t, app, "/tasks")
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter})
	tasksView := e2ePlainView(app)
	if !strings.Contains(tasksView, "drawer: tasks panel: /tasks") {
		t.Fatalf("expected /tasks to open panel after status run")
	}
	if strings.Contains(tasksView, "TASKS_LIST") || strings.Contains(tasksView, "TASKS_STATUS") {
		t.Fatalf("expected no legacy tasks dump while panelized, got:\n%s", tasksView)
	}
}

func TestE2ECommandPanelFlowModelStageThenApplyKeepsSendChainCoherent(t *testing.T) {
	app := e2eReadyApp(t, 180, 30)
	app = e2eTypeText(t, app, "/model")
	app = e2eSendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter})

	stagedView := e2ePlainView(app)
	if !strings.Contains(stagedView, "model picker: /model") {
		t.Fatalf("expected /model to stage picker on first enter")
	}
	if strings.Contains(stagedView, "MODEL_REPORT") {
		t.Fatalf("expected no legacy model contract while picker is open, got:\n%s", stagedView)
	}

	app = e2eSendKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEnter})
	appliedView := e2ePlainView(app)
	if strings.Contains(appliedView, "model picker: /model") {
		t.Fatalf("expected picker closed after apply")
	}
	if !strings.Contains(appliedView, "Model set to") {
		t.Fatalf("expected model confirmation after apply, got:\n%s", appliedView)
	}
	if !strings.Contains(appliedView, "input mode: chat") {
		t.Fatalf("expected chat mode after model apply")
	}
}

func e2eReadyApp(t *testing.T, width, height int) *tui.App {
	t.Helper()
	app := tui.New(tui.Config{})
	updated, _ := app.Update(tea.WindowSizeMsg{Width: width, Height: height})
	next, ok := updated.(*tui.App)
	if !ok {
		t.Fatalf("window size update returned %T, want *tui.App", updated)
	}
	return next
}

func e2eTypeText(t *testing.T, app *tui.App, value string) *tui.App {
	t.Helper()
	next := app
	for _, r := range []rune(value) {
		updated, _ := next.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		cast, ok := updated.(*tui.App)
		if !ok {
			t.Fatalf("typing update returned %T, want *tui.App", updated)
		}
		next = cast
	}
	return next
}

func e2eSendKey(t *testing.T, app *tui.App, msg tea.KeyMsg) *tui.App {
	t.Helper()
	updated, _ := app.Update(msg)
	next, ok := updated.(*tui.App)
	if !ok {
		t.Fatalf("key update returned %T, want *tui.App", updated)
	}
	return next
}

func e2eSendKeyAndRun(t *testing.T, app *tui.App, msg tea.KeyMsg) *tui.App {
	t.Helper()
	updated, cmd := app.Update(msg)
	next, ok := updated.(*tui.App)
	if !ok {
		t.Fatalf("key update returned %T, want *tui.App", updated)
	}
	if cmd == nil {
		return next
	}
	event := cmd()
	if event == nil {
		return next
	}
	updated, _ = next.Update(event)
	next, ok = updated.(*tui.App)
	if !ok {
		t.Fatalf("follow-up update returned %T, want *tui.App", updated)
	}
	return next
}

func e2ePlainView(app *tui.App) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(app.View(), "")
}
