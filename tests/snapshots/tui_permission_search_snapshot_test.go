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
	const wantPrefix = "search: (type to filter)\nmode:timeline  matches:0  jump:ctrl+n/ctrl+p up/down tab shift+tab"
	if !strings.HasPrefix(got, wantPrefix) {
		t.Fatalf("timeline search snapshot mismatch\n--- got ---\n%s\n--- want prefix ---\n%s", got, wantPrefix)
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

func TestSnapshot_QuickOpenFooterShowsActiveRowAndControls(t *testing.T) {
	app := readyApp(t, 180, 26)

	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	searching, ok := updated.(*tui.App)
	if !ok {
		t.Fatalf("ctrl+o update returned %T, want *tui.App", updated)
	}

	plain := stripANSI(searching.View())
	footer := mustFindLineContaining(t, plain, "active: Search timeline")

	got := trimRight(footer)
	const want = "active: Search timeline  |  enter/right apply  left/esc close  tab/down next  shift+tab/up prev  pgup/pgdown page  home/end jump"
	if got != want {
		t.Fatalf("quick-open footer snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSnapshot_PermissionModeStatusBarAfterCommand(t *testing.T) {
	app := readyApp(t, 90, 24)
	app = submitText(t, app, "/permissions auto")

	plain := stripANSI(app.View())
	status := mustFindLineContaining(t, plain, "input chat")
	assistant := mustFindLineContaining(t, plain, "Mode: auto")

	if strings.TrimSpace(assistant) != "Mode: auto" {
		t.Fatalf("expected permission assistant summary, got %q", assistant)
	}
	if !strings.Contains(status, "mode:auto") || !strings.Contains(status, "input chat") {
		t.Fatalf("expected updated status bar with auto mode and chat input, got %q", status)
	}
}
