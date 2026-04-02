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
	runtime := mustFindLineContaining(t, plain, "runtime: state=search")
	context := mustFindLineContaining(t, plain, "context: search mode quick-open")

	got := strings.Join([]string{strings.TrimSpace(runtime), strings.TrimSpace(context)}, "\n")
	want := "runtime: state=search transitions=1 tools=0 permissions=0\ncontext: search mode quick-open selection=Search timeline"
	if got != want {
		t.Fatalf("runtime/context snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}
