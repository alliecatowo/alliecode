package tui

import (
	"strings"
	"testing"
)

func TestQuickOpenStateFiltersAndSelection(t *testing.T) {
	state := newQuickOpenState([]quickOpenItem{
		{label: "Open Config", keywords: "settings preferences", value: "config"},
		{label: "Open Logs", keywords: "history trace", value: "logs"},
		{label: "Restart Service", keywords: "daemon", value: "restart"},
	})

	if len(state.visible) != 3 {
		t.Fatalf("expected all quick-open items visible before query, got %d", len(state.visible))
	}
	if state.selected != 0 {
		t.Fatalf("expected first item selected before query, got %d", state.selected)
	}

	state.setQuery("hist")
	if len(state.visible) != 1 || state.visible[0] != 1 {
		t.Fatalf("expected only logs item after filter, got %#v", state.visible)
	}
	item, ok := state.selectedItem()
	if !ok || item.value != "logs" {
		t.Fatalf("expected selected logs item, got ok=%t value=%q", ok, item.value)
	}

	state.setQuery("missing")
	if len(state.visible) != 0 {
		t.Fatalf("expected no quick-open matches, got %d", len(state.visible))
	}
	if state.selected != -1 {
		t.Fatalf("expected no active selection when empty, got %d", state.selected)
	}
}

func TestQuickOpenStateSupportsSubsequenceMatch(t *testing.T) {
	state := newQuickOpenState([]quickOpenItem{
		{label: "Change model", keywords: "switch model", value: "command.model"},
	})

	state.setQuery("cmdmdl")
	if len(state.visible) != 1 {
		t.Fatalf("expected subsequence query to match quick-open item, got %d matches", len(state.visible))
	}
}

func TestQuickOpenStatePreservesSelectionAcrossQueryRefinement(t *testing.T) {
	state := newQuickOpenState([]quickOpenItem{
		{label: "Search timeline", keywords: "find filter history", value: "search.timeline"},
		{label: "Change model", keywords: "switch model", value: "command.model"},
		{label: "Open logs", keywords: "history trace", value: "logs"},
	})

	state.setQuery("model")
	item, ok := state.selectedItem()
	if !ok || item.value != "command.model" {
		t.Fatalf("expected preselected command.model item, got ok=%t value=%q", ok, item.value)
	}

	state.setQuery("model")
	if state.selected != 0 {
		t.Fatalf("expected selected index remapped within filtered results, got %d", state.selected)
	}
	item, ok = state.selectedItem()
	if !ok || item.value != "command.model" {
		t.Fatalf("expected command.model to remain selected, got ok=%t value=%q", ok, item.value)
	}
}

func TestQuickOpenMoveSelectionWraps(t *testing.T) {
	state := newQuickOpenState([]quickOpenItem{
		{label: "Search timeline", value: "search.timeline"},
		{label: "Change model", value: "command.model"},
	})

	if state.selected != 0 {
		t.Fatalf("expected initial quick-open selection at zero, got %d", state.selected)
	}
	state.moveSelection(-1)
	if state.selected != 1 {
		t.Fatalf("expected reverse wrap to last item, got %d", state.selected)
	}
	state.moveSelection(1)
	if state.selected != 0 {
		t.Fatalf("expected forward wrap to first item, got %d", state.selected)
	}
}

func TestHistorySearchRankingAndRecency(t *testing.T) {
	state := newHistorySearchState([]historySearchEntry{
		{text: "open dashboard"},
		{text: "ship release"},
		{text: "prepare release notes"},
	})

	state.setQuery("release")
	if len(state.matches) != 2 {
		t.Fatalf("expected two matching history entries, got %d", len(state.matches))
	}
	if state.matches[0].index != 2 {
		t.Fatalf("expected newer partial match to rank first, got index %d", state.matches[0].index)
	}

	state.setQuery("ship release")
	if len(state.matches) != 2 {
		t.Fatalf("expected two token matches, got %#v", state.matches)
	}
	if state.matches[0].index != 1 {
		t.Fatalf("expected exact match to rank first, got %#v", state.matches)
	}
	selected, ok := state.selectedEntry()
	if !ok || selected.text != "prepare release notes" {
		t.Fatalf("expected selection to preserve prior focused entry, got ok=%t text=%q", ok, selected.text)
	}
}

func TestHistorySearchMoveSelectionWraps(t *testing.T) {
	state := newHistorySearchState([]historySearchEntry{{text: "alpha"}, {text: "beta"}, {text: "gamma"}})
	state.setQuery("a")
	if len(state.matches) != 3 {
		t.Fatalf("expected three matches, got %d", len(state.matches))
	}

	start := state.selected
	state.moveSelection(1)
	expectedForward := nextMatchPos(start, len(state.matches), 1)
	if state.selected != expectedForward {
		t.Fatalf("expected forward move to %d, got %d", expectedForward, state.selected)
	}

	state.selected = len(state.matches) - 1
	state.moveSelection(1)
	if state.selected != 0 {
		t.Fatalf("expected forward wrap to first index, got %d", state.selected)
	}

	state.selected = 0
	state.moveSelection(-1)
	if state.selected != len(state.matches)-1 {
		t.Fatalf("expected reverse wrap to last index, got %d", state.selected)
	}
}

func TestHistorySearchPreservesSelectionAcrossQueryRefinement(t *testing.T) {
	state := newHistorySearchState([]historySearchEntry{
		{text: "deploy release"},
		{text: "open logs"},
		{text: "ship release"},
	})

	state.setQuery("release")
	if len(state.matches) != 2 {
		t.Fatalf("expected two release matches, got %d", len(state.matches))
	}
	state.moveSelection(1)
	selected, ok := state.selectedEntry()
	if !ok || selected.text != "deploy release" {
		t.Fatalf("expected selected entry deploy release before refinement, got ok=%t text=%q", ok, selected.text)
	}

	state.setQuery("deploy")
	if len(state.matches) != 1 {
		t.Fatalf("expected one deploy match, got %d", len(state.matches))
	}
	if state.selected != 0 {
		t.Fatalf("expected selected index remapped to first match, got %d", state.selected)
	}
	selected, ok = state.selectedEntry()
	if !ok || selected.text != "deploy release" {
		t.Fatalf("expected deploy release to remain selected after refinement, got ok=%t text=%q", ok, selected.text)
	}
}

func TestQuickOpenRenderLinesIncludesDetailsAndSelectionMarker(t *testing.T) {
	state := newQuickOpenState([]quickOpenItem{
		{label: "Search timeline", detail: "Filter visible rows", status: "navigation", keywords: "find", value: "search.timeline", hint: "ctrl+f"},
		{label: "Change model", detail: "Stage /model", status: "command", keywords: "switch", value: "command.model", hint: "/model"},
	})

	state.setQuery("")
	lines := state.renderLines(120, 5)
	if len(lines) != 4 {
		t.Fatalf("expected section headers + 2 quick-open lines, got %d", len(lines))
	}
	if lines[0] != "  Navigation:" {
		t.Fatalf("expected first section header, got %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "> ") {
		t.Fatalf("expected selected quick-open line to use '> ' prefix, got %q", lines[1])
	}
	if !containsAll(lines[1], []string{"Search timeline", "Filter visible rows", "[navigation, ctrl+f]"}) {
		t.Fatalf("expected quick-open line to include label/detail/status, got %q", lines[1])
	}
	if lines[2] != "  Commands:" {
		t.Fatalf("expected second section header, got %q", lines[2])
	}
}

func TestHistoryRenderLinesIncludesScoreAndOverflow(t *testing.T) {
	state := newHistorySearchState([]historySearchEntry{
		{text: "deploy release"},
		{text: "open logs\nfor service A"},
		{text: "restart daemon"},
	})

	state.setQuery("e")
	state.moveSelection(1)
	lines := state.renderLines(120, 2)
	if len(lines) != 3 {
		t.Fatalf("expected 2 lines + overflow indicator, got %d", len(lines))
	}
	if !strings.HasPrefix(lines[1], "> ") {
		t.Fatalf("expected selected history line to use '> ' prefix, got %q", lines[1])
	}
	if !containsAll(lines[0], []string{"[score:"}) {
		t.Fatalf("expected history line to include score status, got %q", lines[0])
	}
	if !strings.HasPrefix(lines[2], "  ... +") {
		t.Fatalf("expected overflow indicator line, got %q", lines[2])
	}
}

func TestHistorySelectedSummaryIncludesScoreAndPosition(t *testing.T) {
	state := newHistorySearchState([]historySearchEntry{{text: "deploy release"}, {text: "open logs"}})
	state.setQuery("release")
	summary := state.selectedSummary(120)
	if !containsAll(summary, []string{"selected: 1/1", "score:", "deploy release"}) {
		t.Fatalf("expected selected summary to include index score and label, got %q", summary)
	}
}

func containsAll(s string, parts []string) bool {
	for _, part := range parts {
		if !strings.Contains(s, part) {
			return false
		}
	}
	return true
}
