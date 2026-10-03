package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alliecatowo/alliecode/internal/references"
)

func (a *App) syncReferenceAutocomplete() {
	if a.stateValue() != stateIdle {
		a.clearReferenceAutocomplete()
		return
	}
	if a.commandPanel.active {
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
	suggestions := a.resolver.Suggest(token.Query, a.recentRefs, a.openRefs, a.contextRefs, 12)
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
	if prevPath == "" {
		prevPath = strings.TrimSpace(strings.ToLower(a.refAuto.last))
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
	a.syncInputMode()
}

func (a *App) clearReferenceAutocomplete() {
	a.rememberReferenceSelection()
	memory := a.refAuto.memory
	last := a.refAuto.last
	a.refAuto = referenceAutocompleteState{selected: -1, offset: 0, memory: memory, last: last}
	if a.refAuto.memory == nil {
		a.refAuto.memory = make(map[string]string, 8)
	}
	a.syncInputMode()
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
	step := 5
	sign := 1
	if dir < 0 {
		sign = -1
	}
	next := a.refAuto.selected + sign*step
	if next < 0 {
		next = 0
	}
	if next >= len(a.refAuto.suggestions) {
		next = len(a.refAuto.suggestions) - 1
	}
	a.refAuto.selected = next
	a.rememberReferenceSelection()
	a.ensureReferenceSelectionVisible(6)
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
	value := strings.TrimSpace(strings.ToLower(a.refAuto.suggestions[a.refAuto.selected].Path))
	a.refAuto.memory[query] = value
	a.refAuto.last = value
}

func (a *App) recallReferenceSelection(query string) string {
	return recallClosestSelection(a.refAuto.memory, query)
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
	a.syncInputMode()
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
	a.recentRefs = mergeRefHistory(next, a.recentRefs, 32)
	a.contextRefs = mergeRefHistory(next, a.contextRefs, 24)
}

func mergeRefHistory(newest []string, prev []string, limit int) []string {
	if limit <= 0 {
		limit = 32
	}
	next := make([]string, 0, limit)
	seen := make(map[string]struct{}, limit)
	for _, path := range newest {
		path = strings.TrimSpace(filepathWithoutLine(path))
		if path == "" {
			continue
		}
		if _, exists := seen[path]; exists {
			continue
		}
		seen[path] = struct{}{}
		next = append(next, path)
		if len(next) >= limit {
			return next
		}
	}
	for _, path := range prev {
		path = strings.TrimSpace(filepathWithoutLine(path))
		if path == "" {
			continue
		}
		if len(next) >= limit {
			break
		}
		if _, exists := seen[path]; exists {
			continue
		}
		seen[path] = struct{}{}
		next = append(next, path)
	}
	return next
}

func (a *App) captureToolReferences(toolName string, raw json.RawMessage) {
	name := strings.ToLower(strings.TrimSpace(toolName))
	if name == "" || len(raw) == 0 {
		return
	}
	if name != "read" && name != "glob" && name != "grep" && name != "edit" && name != "write" && name != "bash" {
		return
	}
	paths := extractPathsFromToolInput(raw)
	if len(paths) == 0 {
		return
	}
	a.openRefs = mergeRefHistory(paths, a.openRefs, 24)
	if name == "read" || name == "edit" || name == "grep" || name == "glob" {
		a.contextRefs = mergeRefHistory(paths, a.contextRefs, 24)
	}
}

func filepathWithoutLine(path string) string {
	if idx := strings.LastIndex(strings.ToUpper(path), "#L"); idx > 0 && idx < len(path)-2 {
		num := strings.TrimSpace(path[idx+2:])
		if dash := strings.Index(num, "-"); dash >= 0 {
			num = strings.TrimSpace(num[:dash])
		}
		allDigits := num != ""
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
	rowWidth := a.width - 2
	if rowWidth < 1 {
		rowWidth = 1
	}
	lines := []string{"references:", panelDivider(rowWidth)}
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
		line := fmt.Sprintf("%s%-36s  [%s %s %s]", prefix, "@"+item.Path, kind, source, reason)
		if item.Exists {
			line += " exists"
		} else {
			line += " missing"
		}
		lines = append(lines, wrapAndClampDisplayLines([]string{line}, rowWidth, 2, "  ...")...)
	}
	if selected, ok := a.selectedReferenceSuggestion(); ok {
		preview := selected.Path
		if selected.IsDir {
			preview += "/"
		}
		displayPath, lineMarker := splitReferenceDisplayPath(selected.Path)
		lines = append(lines, "")
		lines = append(lines, wrapAndClampDisplayLines([]string{"preview: @" + preview}, rowWidth, 2, "  ...")...)
		lines = append(lines, wrapAndClampDisplayLines([]string{"path: " + displayPath}, rowWidth, 2, "  ...")...)
		if lineMarker != "" {
			lines = append(lines, truncateDisplayWidth("line: "+lineMarker, rowWidth, "..."))
		}
		exists := "missing"
		if selected.Exists {
			exists = "exists"
		}
		meta := fmt.Sprintf("meta: score:%d source:%s section:%s reason:%s %s", selected.Score, selected.Source, selected.Section, selected.MatchReason, exists)
		lines = append(lines, wrapAndClampDisplayLines([]string{meta}, rowWidth, 2, "  ...")...)
		if text := strings.TrimSpace(selected.Preview); text != "" {
			lines = append(lines, wrapAndClampDisplayLines([]string{"about: " + text}, rowWidth, 3, "  ...")...)
		}
		lines = append(lines, truncateDisplayWidth(fmt.Sprintf("selection: %d/%d", a.refAuto.selected+1, len(a.refAuto.suggestions)), rowWidth, "..."))
	}
	if len(a.refAuto.suggestions) > window {
		remaining := len(a.refAuto.suggestions) - end
		if remaining > 0 {
			lines = append(lines, fmt.Sprintf("  ... +%d more", remaining))
		}
	}
	lines = append(lines, "")
	lines = append(lines, wrapAndClampDisplayLines([]string{drawerNavHelpSlashRef}, rowWidth, 3, "  ...")...)
	return referenceAutocompleteStyle.Render(strings.Join(lines, "\n"))
}

func extractPathsFromToolInput(raw json.RawMessage) []string {
	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil
	}
	paths := make([]string, 0, 8)
	seen := make(map[string]struct{}, 8)
	var walk func(v any)
	walk = func(v any) {
		switch val := v.(type) {
		case map[string]any:
			for k, child := range val {
				kn := strings.ToLower(strings.TrimSpace(k))
				if (strings.Contains(kn, "path") || strings.Contains(kn, "file") || strings.Contains(kn, "dir") || kn == "cwd") && isLikelyPathString(child) {
					path := strings.TrimSpace(child.(string))
					if _, ok := seen[path]; !ok {
						seen[path] = struct{}{}
						paths = append(paths, path)
					}
				}
				walk(child)
			}
		case []any:
			for _, child := range val {
				walk(child)
			}
		}
	}
	walk(payload)
	return paths
}

func isLikelyPathString(v any) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if strings.ContainsAny(s, "/\\") || strings.HasPrefix(s, ".") || strings.HasPrefix(s, "~") {
		return true
	}
	return strings.Contains(s, ".")
}

func splitReferenceDisplayPath(path string) (string, string) {
	trimmed := strings.TrimSpace(path)
	if idx := strings.LastIndex(strings.ToUpper(trimmed), "#L"); idx > 0 {
		return strings.TrimSpace(trimmed[:idx]), strings.TrimSpace(trimmed[idx:])
	}
	if idx := strings.LastIndex(trimmed, ":"); idx > 0 && idx < len(trimmed)-1 {
		line := strings.TrimSpace(trimmed[idx+1:])
		onlyDigits := line != ""
		for _, r := range line {
			if r < '0' || r > '9' {
				onlyDigits = false
				break
			}
		}
		if onlyDigits {
			return strings.TrimSpace(trimmed[:idx]), "#L" + line
		}
	}
	return trimmed, ""
}
