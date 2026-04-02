package snapshots_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/alliecatowo/alliecode/internal/tui"
)

type tuiScenarioCase struct {
	Name         string   `json:"name"`
	NetworkOnly  bool     `json:"network_only"`
	Width        int      `json:"width"`
	Height       int      `json:"height"`
	Action       string   `json:"action"`
	Command      string   `json:"command"`
	Extract      []string `json:"extract"`
	WantJoined   string   `json:"want_joined"`
	WantContains []string `json:"want_contains"`
}

func TestScenarioMatrix_TUIInteractionStates(t *testing.T) {
	t.Parallel()

	paths := loadTUIMatrixPaths(t)
	for _, path := range paths {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()

			var tc tuiScenarioCase
			decodeTUIMatrixCase(t, path, &tc)
			if tc.NetworkOnly {
				t.Skip("network-only scenario")
			}

			app := readyApp(t, tc.Width, tc.Height)
			app = runTUIAction(t, app, tc)

			plain := stripANSI(app.View())
			for _, needle := range tc.WantContains {
				if !strings.Contains(plain, needle) {
					t.Fatalf("view missing %q\n--- view ---\n%s", needle, plain)
				}
			}

			if len(tc.Extract) > 0 {
				lines := make([]string, 0, len(tc.Extract))
				for _, needle := range tc.Extract {
					line := mustFindLineContaining(t, plain, needle)
					lines = append(lines, trimRight(line))
				}
				if got := strings.Join(lines, "\n"); got != tc.WantJoined {
					t.Fatalf("snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", got, tc.WantJoined)
				}
			}
		})
	}
}

func runTUIAction(t *testing.T, app *tui.App, tc tuiScenarioCase) *tui.App {
	t.Helper()
	switch tc.Action {
	case "ctrl_f":
		updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlF})
		next, ok := updated.(*tui.App)
		if !ok {
			t.Fatalf("ctrl+f update returned %T, want *tui.App", updated)
		}
		return next
	case "ctrl_o":
		updated, _ := app.Update(tea.KeyMsg{Type: tea.KeyCtrlO})
		next, ok := updated.(*tui.App)
		if !ok {
			t.Fatalf("ctrl+o update returned %T, want *tui.App", updated)
		}
		return next
	case "submit_command":
		return submitText(t, app, tc.Command)
	default:
		t.Fatalf("unknown action %q", tc.Action)
		return app
	}
}

func loadTUIMatrixPaths(t *testing.T) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join("testdata", "scenario_matrix", "tui_interaction_state", "*.json"))
	if err != nil {
		t.Fatalf("Glob() error = %v", err)
	}
	if len(paths) == 0 {
		t.Fatalf("no TUI matrix fixtures found")
	}
	sort.Strings(paths)
	return paths
}

func decodeTUIMatrixCase(t *testing.T, path string, out any) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	if err := json.Unmarshal(b, out); err != nil {
		t.Fatalf("Unmarshal(%s) error = %v", path, err)
	}
}
