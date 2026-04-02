package tui

import (
	"encoding/json"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPermissionPromptDescriptionToolTypes(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		input    string
		want     string
	}{
		{name: "bash command", toolName: "Bash", input: `{"command":"git status"}`, want: "execute shell command: git status"},
		{name: "file path", toolName: "Write", input: `{"file_path":"/tmp/a.txt"}`, want: "access local files at: /tmp/a.txt"},
		{name: "webfetch url", toolName: "WebFetch", input: `{"url":"https://example.com"}`, want: "fetch content from: https://example.com"},
		{name: "generic fallback", toolName: "Skill", input: `{}`, want: "run Skill"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := permissionPromptDescription(tc.toolName, json.RawMessage(tc.input))
			if got != tc.want {
				t.Fatalf("unexpected prompt description for %s: got %q want %q", tc.toolName, got, tc.want)
			}
		})
	}
}

func TestPermissionModelSelectionTransitionsAndConfirm(t *testing.T) {
	model := NewPermission("bash", "execute shell command: go test ./...")

	if model.selected != PermissionYes {
		t.Fatalf("expected default selected action to be yes, got %d", model.selected)
	}

	next, cmd := model.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd != nil {
		t.Fatalf("expected tab to only move selection and not emit command")
	}
	if next.selected != PermissionNo {
		t.Fatalf("expected tab to select deny action, got %d", next.selected)
	}

	next, _ = next.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if next.decision != PermissionNo {
		t.Fatalf("expected enter to confirm selected deny action, got %d", next.decision)
	}
}

func TestPermissionModelViewIncludesToolHintsAndSelectedAction(t *testing.T) {
	model := NewPermission("webfetch", "fetch content from: https://example.com")
	model.SetQueueIndex(2, 3)
	view := model.View()

	if !strings.Contains(view, "hint: verify destination host before allowing") {
		t.Fatalf("expected webfetch hint in prompt view, got %q", view)
	}
	if !strings.Contains(view, "risk: medium") {
		t.Fatalf("expected risk classification in prompt view, got %q", view)
	}
	if !strings.Contains(view, "> [y] allow once") {
		t.Fatalf("expected selected action marker in prompt view, got %q", view)
	}
	if !strings.Contains(view, "navigate: tab/shift+tab arrows") {
		t.Fatalf("expected navigation hint in prompt view, got %q", view)
	}
	if !strings.Contains(view, "home/end") {
		t.Fatalf("expected home/end navigation hint, got %q", view)
	}
	if !strings.Contains(view, "queue: 2/3") {
		t.Fatalf("expected queue position in prompt status, got %q", view)
	}
}

func TestPermissionModelHomeEndSelectionShortcuts(t *testing.T) {
	model := NewPermission("bash", "execute shell command")
	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if next.selected != PermissionAlways {
		t.Fatalf("expected end to jump to always selection, got %d", next.selected)
	}
	next, _ = next.Update(tea.KeyMsg{Type: tea.KeyHome})
	if next.selected != PermissionYes {
		t.Fatalf("expected home to jump to allow selection, got %d", next.selected)
	}
}

func TestPermissionModelQueuePreviewRendersNextActions(t *testing.T) {
	model := NewPermission("bash", "execute shell command: go test ./...")
	model.SetQueueIndex(1, 3)
	model.SetQueuePreview([]string{"read: access local files at: internal/tui/app.go", "webfetch: fetch content from: https://example.com"})

	view := model.View()
	if !strings.Contains(view, "queue: 1/3 (2 waiting)") {
		t.Fatalf("expected queue waiting count in view, got %q", view)
	}
	if !strings.Contains(view, "next in queue:") {
		t.Fatalf("expected queue heading in view, got %q", view)
	}
	if !strings.Contains(view, "- read: access local files") {
		t.Fatalf("expected first queue preview action in view, got %q", view)
	}
	if !strings.Contains(view, "webfetch: fetch content") {
		t.Fatalf("expected second queue preview action in view, got %q", view)
	}
}

func TestPermissionModelCtrlNavigationAndStackPreview(t *testing.T) {
	model := NewPermission("bash", "execute shell command")
	model.SetQueueStack([]string{"[active] bash turn 3", "[next] read turn 3"})
	model.SetRecentDecisions([]string{"bash -> deny (turn 2)"})

	next, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlN})
	if next.selected != PermissionNo {
		t.Fatalf("expected ctrl+n to cycle selection, got %d", next.selected)
	}
	next, _ = next.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	if next.selected != PermissionYes {
		t.Fatalf("expected ctrl+p to cycle back, got %d", next.selected)
	}
	view := next.View()
	if !strings.Contains(view, "stack:") || !strings.Contains(view, "recent decisions:") {
		t.Fatalf("expected stack and recent panes in view, got %q", view)
	}
}
