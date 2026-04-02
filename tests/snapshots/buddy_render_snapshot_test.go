package snapshots_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/tui"
)

func TestSnapshot_BuddyRenderNarrow(t *testing.T) {
	app := readyApp(t, 40, 20)
	app = submitText(t, app, "/buddy hatch")

	plain := stripANSI(app.View())
	sprite := mustFindLineContaining(t, plain, "\"hi hi\"")
	bubble := mustFindLineContaining(t, plain, "buddy: hi hi")

	got := trimRight(sprite) + "\n" + trimRight(bubble)
	const want = "<\u00b0~\u00b0> \"hi hi\"\nbuddy: hi hi"
	if got != want {
		t.Fatalf("buddy narrow snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestSnapshot_BuddyRenderFull(t *testing.T) {
	app := readyApp(t, 100, 24)
	app = submitText(t, app, "/buddy hatch")

	plain := stripANSI(app.View())
	bubbleIdx := mustFindLineIndexContaining(t, plain, "buddy: hi hi")
	lines := strings.Split(plain, "\n")
	if bubbleIdx < 5 {
		t.Fatalf("expected at least 5 sprite lines before bubble, got index=%d", bubbleIdx)
	}

	got := strings.Join([]string{
		trimRight(lines[bubbleIdx-5]),
		trimRight(lines[bubbleIdx-4]),
		trimRight(lines[bubbleIdx-3]),
		trimRight(lines[bubbleIdx-2]),
		trimRight(lines[bubbleIdx-1]),
		trimRight(lines[bubbleIdx]),
	}, "\n")
	wantA := "  /^\\  /^\\\n <  \u00b0  \u00b0  >\n (   ~~   )\n  `-vvvv-`\n  buddy\nbuddy: hi hi"
	wantB := "  /^\\  /^\\\n <  \u00b0  \u00b0  >\n (        )\n  `-vvvv-`\n  buddy\nbuddy: hi hi"
	if got != wantA && got != wantB {
		t.Fatalf("buddy full snapshot mismatch\n--- got ---\n%s\n--- want A ---\n%s\n--- want B ---\n%s", got, wantA, wantB)
	}
}

func TestSnapshot_BuddyBubbleHiddenWhenTooNarrow(t *testing.T) {
	app := readyApp(t, 20, 20)
	app = submitText(t, app, "/buddy hatch")

	plain := stripANSI(app.View())
	spriteIdx := mustFindLineIndexContaining(t, plain, "\"hi hi\"")
	lines := strings.Split(plain, "\n")
	if spriteIdx+1 >= len(lines) {
		t.Fatalf("expected line after buddy sprite")
	}

	got := trimRight(lines[spriteIdx]) + "\n" + trimRight(lines[spriteIdx+1])
	const want = "<\u00b0~\u00b0> \"hi hi\"\n"
	if got != want {
		t.Fatalf("buddy hidden bubble snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func readyApp(t *testing.T, width, height int) *tui.App {
	t.Helper()
	app := tui.New(tui.Config{Version: "test"})
	updated, _ := app.Update(tea.WindowSizeMsg{Width: width, Height: height})
	ready, ok := updated.(*tui.App)
	if !ok {
		t.Fatalf("update returned %T, want *tui.App", updated)
	}
	return ready
}

func submitText(t *testing.T, app *tui.App, text string) *tui.App {
	t.Helper()
	updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(text)})
	current, ok := updated.(*tui.App)
	if !ok {
		t.Fatalf("runes update returned %T, want *tui.App", updated)
	}

	updated, cmd := current.Update(tea.KeyMsg{Type: tea.KeyEnter})
	current, ok = updated.(*tui.App)
	if !ok {
		t.Fatalf("enter update returned %T, want *tui.App", updated)
	}
	if cmd == nil {
		return current
	}
	msg := cmd()
	if msg == nil {
		return current
	}

	updated, _ = current.Update(msg)
	finalApp, ok := updated.(*tui.App)
	if !ok {
		t.Fatalf("submit update returned %T, want *tui.App", updated)
	}
	return finalApp
}

func mustFindLineContaining(t *testing.T, full, needle string) string {
	t.Helper()
	for _, line := range strings.Split(full, "\n") {
		if strings.Contains(line, needle) {
			return line
		}
	}
	t.Fatalf("could not find line containing %q in view:\n%s", needle, full)
	return ""
}

func mustFindLineIndexContaining(t *testing.T, full, needle string) int {
	t.Helper()
	for i, line := range strings.Split(full, "\n") {
		if strings.Contains(line, needle) {
			return i
		}
	}
	t.Fatalf("could not find line containing %q in view:\n%s", needle, full)
	return -1
}

func trimRight(s string) string {
	return strings.TrimRight(s, " ")
}
