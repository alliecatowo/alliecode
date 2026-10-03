package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestAppSlashAutocompleteOpenAndFilter(t *testing.T) {
	app := New(Config{})

	for _, r := range []rune("/mo") {
		_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	if !app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash autocomplete to open for slash input")
	}
	if app.slashAutocomplete.query != "mo" {
		t.Fatalf("expected query mo, got %q", app.slashAutocomplete.query)
	}
	if len(app.slashAutocomplete.items) == 0 {
		t.Fatalf("expected filtered command suggestions")
	}
	for _, item := range app.slashAutocomplete.items {
		if strings.TrimSpace(item.MatchReason) == "" {
			t.Fatalf("expected suggestions to include match reason for %q", item.Name)
		}
		if item.MatchReason == "browse" {
			t.Fatalf("expected filtered query to produce non-browse match reason, got %q for %q", item.MatchReason, item.Name)
		}
		if strings.TrimSpace(item.Description) == "" {
			t.Fatalf("expected suggestions to include description for %q", item.Name)
		}
	}
}

func TestAppSlashAutocompleteKeyboardSelectAndApply(t *testing.T) {
	app := New(Config{})

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash autocomplete to be visible after '/'")
	}

	if len(app.slashAutocomplete.items) < 2 {
		t.Fatalf("expected at least two commands for selection test")
	}

	first := app.slashAutocomplete.items[0].Name
	second := app.slashAutocomplete.items[1].Name

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyDown})
	if app.slashAutocomplete.selected != 1 {
		t.Fatalf("expected down key to move selected index to 1, got %d", app.slashAutocomplete.selected)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash autocomplete to close after applying selection")
	}
	if got := app.input.Value(); got != "/"+second {
		if got != "/"+second+" " {
			t.Fatalf("expected selected command %q to be applied, got %q (first was %q)", second, got, first)
		}
	}
}

func TestAppSlashAutocompleteRightArrowAppliesSelection(t *testing.T) {
	app := New(Config{})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyDown})
	selected := app.slashAutocomplete.items[1].Name
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRight})
	if app.slashAutocomplete.isVisible() {
		t.Fatalf("expected right arrow to close slash autocomplete")
	}
	if got := app.input.Value(); got != "/"+selected && got != "/"+selected+" " {
		t.Fatalf("expected right arrow to apply %q, got %q", selected, got)
	}
}

func TestAppSlashAutocompleteEscAndPaging(t *testing.T) {
	app := New(Config{})

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if !app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash autocomplete to be visible after '/'")
	}
	if len(app.slashAutocomplete.items) < 6 {
		t.Fatalf("expected enough slash suggestions for paging")
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if app.slashAutocomplete.selected != 5 {
		t.Fatalf("expected pgdown to advance 5 rows, got %d", app.slashAutocomplete.selected)
	}
	if app.slashAutocomplete.offset != 0 {
		t.Fatalf("expected first page to keep offset at 0, got %d", app.slashAutocomplete.offset)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyHome})
	if app.slashAutocomplete.selected != 0 {
		t.Fatalf("expected home to jump to first suggestion, got %d", app.slashAutocomplete.selected)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if app.slashAutocomplete.selected != len(app.slashAutocomplete.items)-1 {
		t.Fatalf("expected end to jump to last suggestion, got %d", app.slashAutocomplete.selected)
	}
	if app.slashAutocomplete.offset <= 0 {
		t.Fatalf("expected end jump to advance viewport offset, got %d", app.slashAutocomplete.offset)
	}

	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if app.slashAutocomplete.isVisible() {
		t.Fatalf("expected esc to close slash autocomplete")
	}
}

func TestAppCommandContextHintShownAfterSlashSelection(t *testing.T) {
	app := readySizedApp(t, 120, 30)
	app = typeTestText(t, app, "/model")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	if got := app.input.Value(); got != "/model " {
		t.Fatalf("expected /model staged, got %q", got)
	}
	if app.commandPanel.active {
		t.Fatalf("expected model command panel inactive after single-enter stage")
	}
	if !app.modelPickerActive() {
		t.Fatalf("expected model picker to open once /model gains trailing space")
	}
}

func TestSlashAutocompleteRendersSectionHeaders(t *testing.T) {
	app := New(Config{})
	app.width = 120
	for _, r := range []rune("/mo") {
		_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	view := app.renderSlashAutocomplete()
	if !strings.Contains(view, "preview: /") {
		t.Fatalf("expected selected row preview line, got %q", view)
	}
	if !strings.Contains(view, "selection:") {
		t.Fatalf("expected selection metadata line, got %q", view)
	}
}

func TestAppSlashAutocompleteHiddenWhenCommandHasArgs(t *testing.T) {
	app := New(Config{})
	for _, r := range []rune("/model gpt") {
		_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash autocomplete hidden once arguments begin")
	}
}

func TestAppSlashAutocompleteOpensAfterLeadingWhitespace(t *testing.T) {
	app := New(Config{})
	for _, r := range []rune("   /mo") {
		_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if !app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash autocomplete to open even with leading whitespace")
	}
	if app.slashAutocomplete.query != "mo" {
		t.Fatalf("expected slash query mo, got %q", app.slashAutocomplete.query)
	}
}

func TestAppSlashAutocompleteEnterImmediateSubmitForSafeCommand(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	updated, _ := app.Update(submitMsg{text: "/status"})
	app = updated.(*App)
	if len(app.timeline) == 0 {
		t.Fatalf("expected status submit to append timeline output")
	}
	if !strings.Contains(app.timeline[len(app.timeline)-1].text, "STATUS_REPORT") {
		t.Fatalf("expected status command output in timeline, got %q", app.timeline[len(app.timeline)-1].text)
	}
}

func TestAppSlashAutocompleteDoctorRunsDefaultPanelActionOnSingleEnter(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/doctor")
	app = sendTestKeyAndRun(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if app.commandPanel.active {
		t.Fatalf("expected exact /doctor enter to apply default panel action")
	}
	if got := app.input.Value(); got != "" {
		t.Fatalf("expected input cleared after doctor default action, got %q", got)
	}
	if len(app.timeline) == 0 || !strings.Contains(app.timeline[len(app.timeline)-1].text, "DOCTOR_REPORT") {
		t.Fatalf("expected immediate doctor output on single enter, got %#v", app.timeline)
	}
}

func TestAppSlashAutocompleteTasksOpensPanelOnSingleEnter(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/tasks")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if !app.commandPanel.active {
		t.Fatalf("expected /tasks to open interactive command panel")
	}
	if app.commandPanel.panel.Command != "tasks" {
		t.Fatalf("expected tasks panel, got %q", app.commandPanel.panel.Command)
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected no immediate submission for panelized /tasks command")
	}
}

func TestAppSlashAutocompleteShiftTabReverseKeepsDrawerOpen(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/")
	if !app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash drawer active")
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyTab}, "tab")
	selected := app.slashAutocomplete.selected
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyShiftTab}, "shift+tab")
	if !app.slashAutocomplete.isVisible() {
		t.Fatalf("expected shift+tab to keep slash drawer open")
	}
	if app.inputMode != inputModeSlash {
		t.Fatalf("expected slash input mode after reverse, got %s", app.inputMode)
	}
	if app.slashAutocomplete.selected == selected {
		t.Fatalf("expected shift+tab reverse to move selection, stayed at %d", app.slashAutocomplete.selected)
	}
}

func TestAppSlashAutocompleteOneEnterPermissionsOpensPanel(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/permissions")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if !app.commandPanel.active {
		t.Fatalf("expected /permissions to open command panel on one enter")
	}
	if app.commandPanel.panel.Command != "permissions" {
		t.Fatalf("expected permissions panel, got %q", app.commandPanel.panel.Command)
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected no immediate timeline submission for panel command, got %d", len(app.timeline))
	}
}

func TestAppSlashAutocompleteOneEnterProviderOpensPanel(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/provider")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if !app.commandPanel.active {
		t.Fatalf("expected /provider to open command panel on one enter")
	}
	if app.commandPanel.panel.Command != "provider" {
		t.Fatalf("expected provider panel, got %q", app.commandPanel.panel.Command)
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected no immediate timeline submission for panel command, got %d", len(app.timeline))
	}
}

func TestAppSlashAutocompleteEnterStagesArgsForArgumentCommand(t *testing.T) {
	app := readySizedApp(t, 120, 30)
	app = typeTestText(t, app, "/model")
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	if app.commandPanel.active {
		t.Fatalf("expected model command panel to stay closed for staged args flow")
	}
	if got := app.input.Value(); got != "/model " {
		t.Fatalf("expected model command staged, got %q", got)
	}
	if !app.modelPickerActive() {
		t.Fatalf("expected model picker to open for staged /model with trailing space")
	}
	if len(app.timeline) != 0 {
		t.Fatalf("expected no immediate submission for arg command, timeline=%d", len(app.timeline))
	}
}

func TestCommandContextHintIncludesUsageForLeafCommand(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.input.SetValue("/status")

	hint := app.renderCommandContextHint()
	if !strings.Contains(hint, "hint: /status") {
		t.Fatalf("expected status usage hint, got %q", hint)
	}
}

func TestCommandContextHintIncludesUsageForArgsCommand(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.input.SetValue("/model")

	hint := app.renderCommandContextHint()
	if !strings.Contains(hint, "hint: /model") {
		t.Fatalf("expected model usage hint, got %q", hint)
	}
}

func TestModelPickerOpensForModelSlashWithoutArgsAndShowsReadiness(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.height = 30
	_, _ = app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	for _, r := range []rune("/model ") {
		_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	if !app.modelPickerActive() {
		t.Fatalf("expected model picker active for /model with no args")
	}
	view := app.renderModelPicker()
	if !strings.Contains(view, "model picker: /model") {
		t.Fatalf("expected model picker header, got %q", view)
	}
	if !strings.Contains(view, "anthropic") || !strings.Contains(view, "ollama") {
		t.Fatalf("expected provider-grouped rows, got %q", view)
	}
	if !strings.Contains(view, "[ready]") || !strings.Contains(view, "[auth]") {
		t.Fatalf("expected readiness markers in picker, got %q", view)
	}
}

func TestModelPickerEnterDispatchesSelectedModelImmediately(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.height = 30
	_, _ = app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	for _, r := range []rune("/model ") {
		_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if !app.modelPickerActive() {
		t.Fatalf("expected model picker active")
	}
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyDown})
	selected, ok := app.selectedModelPickerItem()
	if !ok {
		t.Fatalf("expected selected model picker item")
	}
	app = sendKeyAndRunCmd(t, app, tea.KeyMsg{Type: tea.KeyEnter}, "enter")

	if app.modelPickerActive() {
		t.Fatalf("expected model picker closed after enter apply")
	}
	if got := app.input.Value(); got != "" {
		t.Fatalf("expected input reset after model picker submit, got %q", got)
	}
	if len(app.timeline) == 0 {
		t.Fatalf("expected immediate model command submission to append timeline")
	}
	last := app.timeline[len(app.timeline)-1].text
	if !strings.Contains(last, "Model set to") || !strings.Contains(last, selected.model.Model) {
		t.Fatalf("expected model set confirmation for %q, got %q", selected.model.Model, last)
	}
}

func TestModelPickerEscDismissesAndReturnsInputFocus(t *testing.T) {
	app := New(Config{})
	app.width = 120
	app.height = 30
	_, _ = app.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	for _, r := range []rune("/model ") {
		_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	if !app.modelPickerActive() {
		t.Fatalf("expected model picker active")
	}
	beforeTimeline := len(app.timeline)
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})

	if app.modelPickerActive() {
		t.Fatalf("expected esc to dismiss model picker")
	}
	if !app.input.focused {
		t.Fatalf("expected input focus restored after esc")
	}
	if len(app.timeline) != beforeTimeline {
		t.Fatalf("expected esc to avoid command submission, timeline=%d want=%d", len(app.timeline), beforeTimeline)
	}
}

func sendKeyAndRunCmd(t *testing.T, app *App, msg tea.KeyMsg, label string) *App {
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

func TestAppSlashAutocompleteShiftTabSequenceVariantKeepsDrawerOpen(t *testing.T) {
	app := readySizedApp(t, 160, 28)
	app = typeTestText(t, app, "/")
	if !app.slashAutocomplete.isVisible() {
		t.Fatalf("expected slash drawer active")
	}
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyTab}, "tab")
	selected := app.slashAutocomplete.selected
	app = sendTestKey(t, app, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'', '[', '1', ';', '2', 'Z'}}, "shift-tab-csi-1-2-z")
	if !app.slashAutocomplete.isVisible() {
		t.Fatalf("expected shift+tab sequence alias to keep slash drawer open")
	}
	if app.inputMode != inputModeSlash {
		t.Fatalf("expected slash input mode after reverse, got %s", app.inputMode)
	}
	if app.slashAutocomplete.selected == selected {
		t.Fatalf("expected shift+tab sequence alias to move selection, stayed at %d", app.slashAutocomplete.selected)
	}
}
