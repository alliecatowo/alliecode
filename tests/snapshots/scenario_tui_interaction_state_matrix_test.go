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
	Actions      []string `json:"actions"`
	Command      string   `json:"command"`
	Extract      []string `json:"extract"`
	WantJoined   string   `json:"want_joined"`
	WantContains []string `json:"want_contains"`
	WantAbsent   []string `json:"want_absent"`
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
			for _, needle := range tc.WantAbsent {
				if strings.Contains(plain, needle) {
					t.Fatalf("view unexpectedly contained %q\n--- view ---\n%s", needle, plain)
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
	if len(tc.Actions) > 0 {
		for _, action := range tc.Actions {
			app = applyTUIActionStep(t, app, action)
		}
		return app
	}
	switch tc.Action {
	case "ctrl_f":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, "ctrl+f")
	case "ctrl_o":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, "ctrl+o")
	case "submit_command":
		return submitText(t, app, tc.Command)
	default:
		t.Fatalf("unknown action %q", tc.Action)
		return app
	}
}

func applyTUIActionStep(t *testing.T, app *tui.App, step string) *tui.App {
	t.Helper()
	if strings.HasPrefix(step, "type:") {
		text := strings.TrimPrefix(step, "type:")
		if text == "" {
			return app
		}
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(text)}, step)
	}
	if strings.HasPrefix(step, "submit:") {
		return submitText(t, app, strings.TrimPrefix(step, "submit:"))
	}
	switch step {
	case "ctrl_f":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlF}, step)
	case "ctrl_o":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlO}, step)
	case "ctrl_r":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlR}, step)
	case "ctrl_n":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlN}, step)
	case "ctrl_p":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyCtrlP}, step)
	case "down":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyDown}, step)
	case "up":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyUp}, step)
	case "right":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyRight}, step)
	case "left":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyLeft}, step)
	case "tab":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyTab}, step)
	case "shift_tab":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, step)
	case "shift_tab_csi_z":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', 'Z'}}, step)
	case "shift_tab_csi_1_2_z":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', '1', ';', '2', 'Z'}}, step)
	case "shift_tab_csi_9_2_u":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', '9', ';', '2', 'u'}}, step)
	case "enter":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, step)
	case "shift_enter_csi_13_2_u":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'\x1b', '[', '1', '3', ';', '2', 'u'}}, step)
	case "esc":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyEsc}, step)
	case "pgdown":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyPgDown}, step)
	case "pgup":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyPgUp}, step)
	case "home":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyHome}, step)
	case "end":
		return sendKey(t, app, tea.KeyMsg{Type: tea.KeyEnd}, step)
	default:
		t.Fatalf("unknown scripted action step %q", step)
		return app
	}
}

func sendKey(t *testing.T, app *tui.App, msg tea.KeyMsg, label string) *tui.App {
	t.Helper()
	updated, _ := app.Update(msg)
	next, ok := updated.(*tui.App)
	if !ok {
		t.Fatalf("%s update returned %T, want *tui.App", label, updated)
	}
	return next
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
