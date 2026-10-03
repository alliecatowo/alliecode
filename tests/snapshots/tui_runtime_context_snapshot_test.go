package snapshots_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/tui"
)

func TestSnapshot_RuntimePaneAndContextHintForQuickOpen(t *testing.T) {
	app := readyApp(t, 160, 24)
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	searching, ok := updated.(*tui.App)
	if !ok {
		t.Fatalf("ctrl+o update returned %T", updated)
	}
	plain := stripANSI(searching.View())
	runtime := mustFindLineContaining(t, plain, "runtime: search  runtime: state=search")
	context := mustFindLineContaining(t, plain, "context: search mode quick-open selection=")

	got := strings.Join([]string{strings.TrimSpace(runtime), strings.TrimSpace(context)}, "\n")
	want := "runtime: search  runtime: state=search transitions=1 tools=0 permissions=0 render=0ms\ncontext: search mode quick-open selection=Search timeline"
	if got != want {
		t.Fatalf("runtime/context snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSnapshot_SearchPanelKeepsInputVisibleWithStableHeight(t *testing.T) {
	app := readyApp(t, 120, 28)

	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
	searching, ok := updated.(*tui.App)
	if !ok {
		t.Fatalf("ctrl+o update returned %T", updated)
	}
	plain := stripANSI(searching.View())
	searchLine := mustFindLineContaining(t, plain, "search: (type to filter)")
	inputLine := mustFindLineContaining(t, plain, "Type a message")

	got := strings.Join([]string{strings.TrimSpace(searchLine), strings.TrimSpace(inputLine)}, "\n")
	want := "search: (type to filter)\n┃ Type a message... (Enter to send, Shift+Enter for newline)"
	if got != want {
		t.Fatalf("search/input visibility snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
