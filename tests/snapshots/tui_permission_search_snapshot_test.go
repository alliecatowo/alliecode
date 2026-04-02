package snapshots_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/tui"
)

func TestSnapshot_TimelineSearchStateChrome(t *testing.T) {
	app := readyApp(t, 80, 24)

	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
	searching, ok := updated.(*tui.App)
	if !ok {
		t.Fatalf("ctrl+f update returned %T, want *tui.App", updated)
	}

	plain := stripANSI(searching.View())
	label := mustFindLineContaining(t, plain, "search: (type to filter)")
	meta := mustFindLineContaining(t, plain, "mode:timeline")

	got := trimRight(label) + "\n" + trimRight(meta)
	const want = "search: (type to filter)\nmode:timeline  matches:0  jump:ctrl+n/ctrl+p  esc:exit"
	if got != want {
		t.Fatalf("timeline search snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSnapshot_QuickOpenShowsKeyHintsAndWorkflowRows(t *testing.T) {
	app := readyApp(t, 180, 26)

	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	searching, ok := updated.(*tui.App)
	if !ok {
		t.Fatalf("ctrl+o update returned %T, want *tui.App", updated)
	}

	plain := stripANSI(searching.View())
	meta := mustFindLineContaining(t, plain, "mode:quick-open")
	keys := mustFindLineContaining(t, plain, "keys:")
	workflow := mustFindLineContaining(t, plain, "selected: 1/10")

	got := trimRight(meta) + "\n" + trimRight(keys) + "\n" + trimRight(workflow)
	const want = "mode:quick-open  matches:10  jump:ctrl+n/ctrl+p up/down tab pgup/pgdown  esc:exit\nkeys: ctrl+f -> search:timeline  |  ctrl+n -> select:next  |  ctrl+o -> search:quick_open  |  ctrl+p -> select:previous\nselected: 1/10  Search timeline - Filter visible conversation rows"
	if got != want {
		t.Fatalf("quick-open snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSnapshot_PermissionModeStatusBarAfterCommand(t *testing.T) {
	app := readyApp(t, 90, 24)
	app = submitText(t, app, "/permissions auto")

	plain := stripANSI(app.View())
	status := mustFindLineContaining(t, plain, "mode:auto")
	assistant := mustFindLineContaining(t, plain, "PERMISSIONS_SET")

	got := strings.Join([]string{trimRight(assistant), trimRight(status)}, "\n")
	const want = "  PERMISSIONS_SET\n model:unknown turns:0 mode:auto render:0ms tools:0 perm:0   budget:$0.0000 tok:0 sid:n/a"
	if got != want {
		t.Fatalf("permission status snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
