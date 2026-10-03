package keybindings

import "strings"

func (s *Set) CompactHintsForModes(modes []Mode, limit int) []string {
	hints := s.HintsForModes(modes, limit)
	out := make([]string, 0, len(hints))
	for _, hint := range hints {
		combo, action, ok := strings.Cut(hint, " -> ")
		if !ok {
			out = append(out, hint)
			continue
		}
		out = append(out, strings.TrimSpace(combo)+" "+compactActionLabel(action))
	}
	return out
}

func compactActionLabel(action string) string {
	action = strings.TrimSpace(action)
	if action == "" {
		return "-"
	}
	if domain, op, ok := strings.Cut(action, ":"); ok {
		domain = strings.TrimSpace(domain)
		op = strings.TrimSpace(op)
		if domain != "" && op != "" {
			return domain + "/" + op
		}
	}
	return action
}
