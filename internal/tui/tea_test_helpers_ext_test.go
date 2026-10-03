package tui

import (
	"regexp"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

var testANSIPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func readySizedApp(t *testing.T, width, height int) *App {
	t.Helper()
	app := New(Config{Version: "test"})
	updated, _ := app.Update(tea.WindowSizeMsg{Width: width, Height: height})
	ready, ok := updated.(*App)
	if !ok {
		t.Fatalf("window update returned %T, want *App", updated)
	}
	return ready
}

func sendTestKey(t *testing.T, app *App, msg tea.KeyMsg, label string) *App {
	t.Helper()
	updated, _ := app.Update(msg)
	next, ok := updated.(*App)
	if !ok {
		t.Fatalf("%s update returned %T, want *App", label, updated)
	}
	return next
}

func sendTestKeyAndRun(t *testing.T, app *App, msg tea.KeyMsg, label string) *App {
	t.Helper()
	updated, cmd := app.Update(msg)
	next, ok := updated.(*App)
	if !ok {
		t.Fatalf("%s update returned %T, want *App", label, updated)
	}
	if cmd == nil {
		return next
	}
	event := cmd()
	if event == nil {
		return next
	}
	updated, _ = next.Update(event)
	next, ok = updated.(*App)
	if !ok {
		t.Fatalf("%s command update returned %T, want *App", label, updated)
	}
	return next
}

func typeTestText(t *testing.T, app *App, text string) *App {
	t.Helper()
	if text == "" {
		return app
	}
	return sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(text)}, "type:"+text)
}

func submitSlashCommand(t *testing.T, app *App, text string) *App {
	t.Helper()
	app = typeTestText(t, app, text)
	return sendTestKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
}

func stripANSIForTest(s string) string {
	return testANSIPattern.ReplaceAllString(s, "")
}
