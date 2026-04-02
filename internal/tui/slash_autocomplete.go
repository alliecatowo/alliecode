package tui

import (
	"strings"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func (s *slashAutocompleteState) isVisible() bool {
	return len(s.items) > 0
}

func (s *slashAutocompleteState) clear() {
	s.rememberSelection()
	s.query = ""
	s.items = nil
	s.selected = -1
	s.offset = 0
}

func (s *slashAutocompleteState) setItems(query string, items []commands.Suggestion) {
	prevQuery := s.query
	prevName := ""
	if item, ok := s.selectedItem(); ok {
		prevName = item.Name
	}
	s.rememberSelection()
	s.query = query
	s.items = append([]commands.Suggestion(nil), items...)
	if len(s.items) == 0 {
		s.selected = -1
		s.offset = 0
		return
	}
	if prevName == "" {
		if remembered := s.recallSelection(query); remembered != "" {
			prevName = remembered
		}
	}
	if prevName == "" && prevQuery != "" && prevQuery != query {
		if remembered := s.recallSelection(prevQuery); remembered != "" {
			prevName = remembered
		}
	}
	if prevName != "" {
		for i, item := range s.items {
			if strings.EqualFold(strings.TrimSpace(item.Name), strings.TrimSpace(prevName)) {
				s.selected = i
				s.ensureVisible(6)
				return
			}
		}
	}
	if s.selected < 0 || s.selected >= len(s.items) {
		s.selected = 0
	}
	s.ensureVisible(6)
}

func (s *slashAutocompleteState) moveSelection(delta int) {
	if len(s.items) == 0 {
		s.selected = -1
		return
	}
	s.selected = nextMatchPos(s.selected, len(s.items), delta)
	s.rememberSelection()
	s.ensureVisible(6)
}

func (s *slashAutocompleteState) pageSelection(dir int) {
	if len(s.items) == 0 {
		s.selected = -1
		return
	}
	for i := 0; i < 5; i++ {
		s.selected = nextMatchPos(s.selected, len(s.items), dir)
	}
	s.rememberSelection()
	s.ensureVisible(6)
}

func (s *slashAutocompleteState) jumpSelection(toEnd bool) {
	if len(s.items) == 0 {
		s.selected = -1
		return
	}
	if toEnd {
		s.selected = len(s.items) - 1
		s.rememberSelection()
		s.ensureVisible(6)
		return
	}
	s.selected = 0
	s.rememberSelection()
	s.ensureVisible(6)
}

func (s *slashAutocompleteState) rememberSelection() {
	if len(s.items) == 0 || s.selected < 0 || s.selected >= len(s.items) {
		return
	}
	query := strings.TrimSpace(strings.ToLower(s.query))
	if query == "" {
		return
	}
	if s.memory == nil {
		s.memory = make(map[string]string, 8)
	}
	s.memory[query] = strings.TrimSpace(strings.ToLower(s.items[s.selected].Name))
}

func (s slashAutocompleteState) recallSelection(query string) string {
	if len(s.memory) == 0 {
		return ""
	}
	return strings.TrimSpace(s.memory[strings.TrimSpace(strings.ToLower(query))])
}

func (s *slashAutocompleteState) ensureVisible(window int) {
	if window <= 0 {
		window = 6
	}
	if s.selected < 0 {
		s.offset = 0
		return
	}
	if s.selected < s.offset {
		s.offset = s.selected
		return
	}
	if s.selected >= s.offset+window {
		s.offset = s.selected - window + 1
	}
	maxOffset := len(s.items) - window
	if maxOffset < 0 {
		maxOffset = 0
	}
	if s.offset > maxOffset {
		s.offset = maxOffset
	}
}

func (s slashAutocompleteState) selectedItem() (commands.Suggestion, bool) {
	if s.selected < 0 || s.selected >= len(s.items) {
		return commands.Suggestion{}, false
	}
	return s.items[s.selected], true
}
