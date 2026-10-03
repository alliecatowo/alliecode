package tui

import (
	"strings"

	"github.com/alliecatowo/alliecode/internal/commands"
)

func (s *commandPanelState) activate(panel commands.InteractivePanel, input string) {
	s.active = true
	s.panel = panel
	s.previousInput = input
	prevKey := ""
	if item, ok := s.selectedItem(); ok {
		prevKey = item.Key
	}
	if prevKey == "" {
		prevKey = s.recallSelection(panel.Command)
	}
	s.selected = 0
	s.offset = 0
	if prevKey != "" {
		for i, item := range panel.Items {
			if strings.EqualFold(strings.TrimSpace(item.Key), strings.TrimSpace(prevKey)) {
				s.selected = i
				break
			}
		}
	}
	s.ensureVisible(6)
}

func (s *commandPanelState) clear() {
	s.rememberSelection()
	s.active = false
	s.panel = commands.InteractivePanel{}
	s.selected = -1
	s.offset = 0
	s.previousInput = ""
}

func (s *commandPanelState) selectedItem() (commands.InteractivePanelItem, bool) {
	if !s.active || s.selected < 0 || s.selected >= len(s.panel.Items) {
		return commands.InteractivePanelItem{}, false
	}
	return s.panel.Items[s.selected], true
}

func (s *commandPanelState) moveSelection(delta int) {
	if len(s.panel.Items) == 0 {
		s.selected = -1
		return
	}
	s.selected = nextMatchPos(s.selected, len(s.panel.Items), delta)
	s.rememberSelection()
	s.ensureVisible(6)
}

func (s *commandPanelState) pageSelection(dir int) {
	if len(s.panel.Items) == 0 {
		s.selected = -1
		return
	}
	step := 5
	sign := 1
	if dir < 0 {
		sign = -1
	}
	next := s.selected + sign*step
	if next < 0 {
		next = 0
	}
	if next >= len(s.panel.Items) {
		next = len(s.panel.Items) - 1
	}
	s.selected = next
	s.rememberSelection()
	s.ensureVisible(6)
}

func (s *commandPanelState) jumpSelection(toEnd bool) {
	if len(s.panel.Items) == 0 {
		s.selected = -1
		return
	}
	if toEnd {
		s.selected = len(s.panel.Items) - 1
	} else {
		s.selected = 0
	}
	s.rememberSelection()
	s.ensureVisible(6)
}

func (s *commandPanelState) ensureVisible(window int) {
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
	maxOffset := len(s.panel.Items) - window
	if maxOffset < 0 {
		maxOffset = 0
	}
	if s.offset > maxOffset {
		s.offset = maxOffset
	}
}

func (s *commandPanelState) rememberSelection() {
	if !s.active || len(s.panel.Items) == 0 || s.selected < 0 || s.selected >= len(s.panel.Items) {
		return
	}
	key := strings.TrimSpace(strings.ToLower(s.panel.Command))
	if key == "" {
		return
	}
	if s.memory == nil {
		s.memory = make(map[string]string, 8)
	}
	s.memory[key] = strings.TrimSpace(strings.ToLower(s.panel.Items[s.selected].Key))
}

func (s *commandPanelState) recallSelection(command string) string {
	if len(s.memory) == 0 {
		return ""
	}
	return strings.TrimSpace(s.memory[strings.TrimSpace(strings.ToLower(command))])
}
