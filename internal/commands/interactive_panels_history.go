package commands

import (
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/types"
)

func buildHistoryInteractivePanel(cmdCtx Context) InteractivePanel {
	state := interactivePanelState(cmdCtx)
	entries := defaultHistoryEntries(state)
	panel := InteractivePanel{
		Command:       "history",
		Title:         "history panel: /history",
		Subtitle:      fmt.Sprintf("entries=%d views=%d", len(entries), state.HistoryViews),
		HeaderIntents: historyStatusIntents(state, len(entries), uniqueHistoryModels(entries)),
		Items: []InteractivePanelItem{
			{Key: "latest", Section: "Overview", Label: "Latest history entry", Detail: "Show the newest stored session", Status: statusWord(len(entries) > 0, "recent", "empty"), ApplyInput: "/history latest", ApplyMode: PanelApplySubmit, PreviewIntents: historyLatestPanelIntents(entries), Preview: previewLines("Use latest to inspect the newest session transcript metadata.")},
			{Key: "list", Section: "Overview", Label: "List recent history", Detail: "Render the full local history list", Status: statusWord(len(entries) > 0, "inventory", "empty"), ApplyInput: "/history list", ApplyMode: PanelApplySubmit, PreviewIntents: historyListIntents("all", "-", len(entries), entries), Preview: previewLines(fmt.Sprintf("Stored history entries: %d", len(entries)))},
			{Key: "status", Section: "Overview", Label: "History status", Detail: "Inspect history filter and view counters", Status: "status", ApplyInput: "/history status", ApplyMode: PanelApplySubmit, PreviewIntents: historyStatusIntents(state, len(entries), uniqueHistoryModels(entries)), Preview: previewLines(fmt.Sprintf("Last filter: %s=%s", normalizeToken(state.HistoryLastFilterType), normalizeToken(state.HistoryLastFilterValue)))},
			{Key: "doctor", Section: "Diagnostics", Label: "History doctor", Detail: "Run deterministic history diagnostics", Status: "diagnostics", ApplyInput: "/history doctor", ApplyMode: PanelApplySubmit, PreviewIntents: historyDoctorIntents(len(entries), "/history list"), Preview: previewLines("History doctor suggests quick fixes when history is empty or filtering is stale.")},
		},
	}

	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		panel.Items = append(panel.Items, InteractivePanelItem{
			Key:            entry.ID,
			Section:        "Entries",
			Label:          fmt.Sprintf("%s (%s)", entry.Title, entry.ID),
			Detail:         fmt.Sprintf("model=%s turns=%d created=%s", entry.Model, entry.Turns, entry.CreatedAt),
			Status:         "entry",
			ApplyInput:     "/history show " + entry.ID,
			ApplyMode:      PanelApplySubmit,
			PreviewIntents: historyShowIntents(entry),
			Preview: previewLines(
				fmt.Sprintf("Path: %s", entry.Path),
				fmt.Sprintf("Model: %s", entry.Model),
				fmt.Sprintf("Turns: %d", entry.Turns),
				fmt.Sprintf("Summary: %s", normalizeToken(entry.Summary)),
			),
		})
	}

	return panel
}

func uniqueHistoryModels(entries []HistoryEntry) int {
	seen := map[string]struct{}{}
	for _, entry := range entries {
		if model := strings.TrimSpace(entry.Model); model != "" {
			seen[model] = struct{}{}
		}
	}
	return len(seen)
}

func historyLatestPanelIntents(entries []HistoryEntry) []types.RenderIntent {
	if len(entries) == 0 {
		return historyLatestIntents(HistoryEntry{}, false)
	}
	return historyLatestIntents(entries[len(entries)-1], true)
}
