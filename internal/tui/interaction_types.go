package tui

import (
	"strings"

	"github.com/alliecatowo/alliecode/internal/commands"
	"github.com/alliecatowo/alliecode/internal/providers"
	"github.com/alliecatowo/alliecode/internal/references"
)

type slashAutocompleteState struct {
	query       string
	items       []commands.Suggestion
	selected    int
	offset      int
	memory      map[string]string
	modelPicker modelPickerState
}

type modelPickerItem struct {
	provider string
	model    providers.ModelMetadata
	ready    bool
}

type modelPickerState struct {
	active          bool
	items           []modelPickerItem
	selected        int
	offset          int
	dismissedInput  string
	showSlashHeader bool
}

type commandPanelState struct {
	active        bool
	panel         commands.InteractivePanel
	selected      int
	offset        int
	memory        map[string]string
	previousInput string
}

type permissionPromptRequest struct {
	queueKey    string
	toolUseID   string
	toolName    string
	toolKind    string
	toolDetails []string
	status      permissionStatus
	turn        int
	description string
	timelineRow int
}

type referenceAutocompleteState struct {
	active      bool
	token       referenceToken
	suggestions []references.Suggestion
	selected    int
	offset      int
	memory      map[string]string
	last        string
}

type timelineSearchMatch struct {
	rowIndex    int
	lineOffset  int
	occurrence  int
	rowMatches  int
	excerpt     string
	matchReason string
}

type permissionDecisionRecord struct {
	ToolName  string
	Decision  PermissionDecision
	Status    permissionStatus
	Turn      int
	QueueKey  string
	Timestamp int64
}

type searchMode int

const (
	searchModeTimeline searchMode = iota
	searchModeQuickOpen
	searchModeHistory
)

func (m searchMode) label() string {
	switch m {
	case searchModeQuickOpen:
		return "quick-open"
	case searchModeHistory:
		return "history-search"
	default:
		return "timeline"
	}
}

func historyEntriesFromInput(history []string) []historySearchEntry {
	entries := make([]historySearchEntry, 0, len(history))
	for _, value := range history {
		if strings.TrimSpace(value) == "" {
			continue
		}
		entries = append(entries, historySearchEntry{text: value})
	}
	return entries
}
