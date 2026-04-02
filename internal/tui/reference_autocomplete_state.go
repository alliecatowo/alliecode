package tui

import (
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/references"
)

func (a *App) syncReferenceAutocomplete() {
	if a.state != stateIdle {
		a.clearReferenceAutocomplete()
		return
	}
	if a.slashAutocomplete.isVisible() {
		a.clearReferenceAutocomplete()
		return
	}
	token, ok := currentReferenceToken(a.input.Value())
	if !ok {
		a.clearReferenceAutocomplete()
		return
	}

	prevPath := ""
	prevQuery := ""
	if item, exists := a.selectedReferenceSuggestion(); exists {
		prevPath = item.Path
		prevQuery = strings.TrimSpace(strings.ToLower(a.refAuto.token.Query))
	}
	a.rememberReferenceSelection()
	suggestions := a.resolver.Suggest(token.Query, a.recentRefs, 8)
	if len(suggestions) == 0 {
		a.clearReferenceAutocomplete()
		return
	}

	a.refAuto.active = true
	a.refAuto.token = token
	a.refAuto.suggestions = suggestions
	a.refAuto.selected = 0
	a.refAuto.offset = 0
	if prevPath == "" {
		prevPath = a.recallReferenceSelection(token.Query)
	}
	if prevPath == "" && prevQuery != "" {
		prevPath = a.recallReferenceSelection(prevQuery)
	}
	if prevPath != "" {
		for i, item := range suggestions {
			if strings.EqualFold(strings.TrimSpace(item.Path), strings.TrimSpace(prevPath)) {
				a.refAuto.selected = i
				a.ensureReferenceSelectionVisible(6)
				break
			}
		}
	}
}

func (a *App) clearReferenceAutocomplete() {
	a.rememberReferenceSelection()
	memory := a.refAuto.memory
	a.refAuto = referenceAutocompleteState{selected: -1, offset: 0, memory: memory}
}

func (a *App) selectedReferenceSuggestion() (references.Suggestion, bool) {
	if !a.refAuto.active || len(a.refAuto.suggestions) == 0 {
		return references.Suggestion{}, false
	}
	if a.refAuto.selected < 0 || a.refAuto.selected >= len(a.refAuto.suggestions) {
		return references.Suggestion{}, false
	}
	return a.refAuto.suggestions[a.refAuto.selected], true
}

func (a *App) advanceReferenceSelection(dir int) {
	if !a.refAuto.active || len(a.refAuto.suggestions) == 0 {
		return
	}
	a.refAuto.selected = nextMatchPos(a.refAuto.selected, len(a.refAuto.suggestions), dir)
	a.rememberReferenceSelection()
	a.ensureReferenceSelectionVisible(6)
}

func (a *App) pageReferenceSelection(dir int) {
	if !a.refAuto.active || len(a.refAuto.suggestions) == 0 {
		return
	}
	for i := 0; i < 5; i++ {
		a.advanceReferenceSelection(dir)
	}
}

func (a *App) jumpReferenceSelection(toEnd bool) {
	if !a.refAuto.active || len(a.refAuto.suggestions) == 0 {
		return
	}
	if toEnd {
		a.refAuto.selected = len(a.refAuto.suggestions) - 1
	} else {
		a.refAuto.selected = 0
	}
	a.rememberReferenceSelection()
	a.ensureReferenceSelectionVisible(6)
}

func (a *App) rememberReferenceSelection() {
	if len(a.refAuto.suggestions) == 0 || a.refAuto.selected < 0 || a.refAuto.selected >= len(a.refAuto.suggestions) {
		return
	}
	query := strings.TrimSpace(strings.ToLower(a.refAuto.token.Query))
	if query == "" {
		return
	}
	if a.refAuto.memory == nil {
		a.refAuto.memory = make(map[string]string, 8)
	}
	a.refAuto.memory[query] = strings.TrimSpace(strings.ToLower(a.refAuto.suggestions[a.refAuto.selected].Path))
}

func (a *App) recallReferenceSelection(query string) string {
	if len(a.refAuto.memory) == 0 {
		return ""
	}
	return strings.TrimSpace(a.refAuto.memory[strings.TrimSpace(strings.ToLower(query))])
}

func (a *App) ensureReferenceSelectionVisible(window int) {
	if window <= 0 {
		window = 6
	}
	if a.refAuto.selected < 0 {
		a.refAuto.offset = 0
		return
	}
	if a.refAuto.selected < a.refAuto.offset {
		a.refAuto.offset = a.refAuto.selected
		return
	}
	if a.refAuto.selected >= a.refAuto.offset+window {
		a.refAuto.offset = a.refAuto.selected - window + 1
	}
	maxOffset := len(a.refAuto.suggestions) - window
	if maxOffset < 0 {
		maxOffset = 0
	}
	if a.refAuto.offset > maxOffset {
		a.refAuto.offset = maxOffset
	}
}

func (a *App) applyReferenceSelection() bool {
	item, ok := a.selectedReferenceSuggestion()
	if !ok {
		return false
	}
	updated := buildReferenceInsertion(a.input.Value(), a.refAuto.token, item)
	a.input.SetValue(updated)
	a.input.CursorEnd()
	a.clearReferenceAutocomplete()
	return true
}

func (a *App) captureRecentReferences(text string) {
	refs := references.ParseReferences(text)
	if len(refs) == 0 {
		return
	}
	next := make([]string, 0, 32)
	seen := make(map[string]struct{}, 32)
	for _, ref := range refs {
		path := strings.TrimSpace(filepathWithoutLine(ref.Path))
		if path == "" {
			continue
		}
		if _, exists := seen[path]; exists {
			continue
		}
		seen[path] = struct{}{}
		next = append(next, path)
	}
	for _, path := range a.recentRefs {
		if len(next) >= 32 {
			break
		}
		if _, exists := seen[path]; exists {
			continue
		}
		seen[path] = struct{}{}
		next = append(next, path)
	}
	a.recentRefs = next
}

func filepathWithoutLine(path string) string {
	if idx := strings.LastIndex(path, ":"); idx > 0 && idx < len(path)-1 {
		num := strings.TrimSpace(path[idx+1:])
		allDigits := true
		for _, r := range num {
			if r < '0' || r > '9' {
				allDigits = false
				break
			}
		}
		if allDigits {
			return path[:idx]
		}
	}
	return path
}

func (a *App) renderReferenceAutocomplete() string {
	if !a.refAuto.active || len(a.refAuto.suggestions) == 0 {
		return ""
	}
	lines := []string{"references:"}
	rowWidth := a.width - 2
	if rowWidth < 1 {
		rowWidth = 1
	}
	window := 6
	start := a.refAuto.offset
	if start < 0 || start >= len(a.refAuto.suggestions) {
		start = 0
	}
	end := start + window
	if end > len(a.refAuto.suggestions) {
		end = len(a.refAuto.suggestions)
	}
	lastSection := ""
	sectionCounts := make(map[string]int, 8)
	for _, item := range a.refAuto.suggestions {
		section := strings.TrimSpace(item.Section)
		if section == "" {
			section = "Workspace"
		}
		sectionCounts[section]++
	}
	for i := start; i < end; i++ {
		item := a.refAuto.suggestions[i]
		section := strings.TrimSpace(item.Section)
		if section == "" {
			section = "Workspace"
		}
		if section != lastSection {
			lines = append(lines, "  "+section+" ("+itoa(sectionCounts[section])+"):")
			lastSection = section
		}
		prefix := "  "
		if i == a.refAuto.selected {
			prefix = "> "
		}
		kind := "file"
		if item.IsDir {
			kind = "dir"
		}
		source := item.Source
		if source == "workspace+recent" {
			source = "recent"
		}
		reason := strings.TrimSpace(item.MatchReason)
		if reason == "" {
			reason = "browse"
		}
		line := fmt.Sprintf("%s@%s [%s %s %s]", prefix, item.Path, kind, source, reason)
		lines = append(lines, truncateDisplayWidth(line, rowWidth, "..."))
	}
	if selected, ok := a.selectedReferenceSuggestion(); ok {
		preview := selected.Path
		if selected.IsDir {
			preview += "/"
		}
		lines = append(lines, "")
		lines = append(lines, truncateDisplayWidth("preview: @"+preview, rowWidth, "..."))
		lines = append(lines, truncateDisplayWidth(fmt.Sprintf("meta: score:%d source:%s reason:%s", selected.Score, selected.Source, selected.MatchReason), rowWidth, "..."))
		if text := strings.TrimSpace(selected.Preview); text != "" {
			lines = append(lines, truncateDisplayWidth("about: "+text, rowWidth, "..."))
		}
	}
	if len(a.refAuto.suggestions) > window {
		remaining := len(a.refAuto.suggestions) - end
		if remaining > 0 {
			lines = append(lines, fmt.Sprintf("  ... +%d more", remaining))
		}
	}
	lines = append(lines, "")
	lines = append(lines, truncateDisplayWidth("tab/enter apply  esc dismiss  up/down navigate  pgup/pgdown page  home/end jump", rowWidth, "..."))
	return referenceAutocompleteStyle.Render(strings.Join(lines, "\n"))
}
