package keybindings

import "strings"

// ModeStack returns an ordered mode stack with optional global fallback.
func ModeStack(primary []Mode, includeGlobal bool) []Mode {
	seen := make(map[Mode]struct{}, len(primary)+1)
	out := make([]Mode, 0, len(primary)+1)
	for _, mode := range primary {
		if _, ok := seen[mode]; ok {
			continue
		}
		seen[mode] = struct{}{}
		out = append(out, mode)
	}
	if includeGlobal {
		if _, ok := seen[ModeGlobal]; !ok {
			out = append(out, ModeGlobal)
		}
	}
	return out
}

// ResolveActionInStack resolves a combo with deterministic mode-layer precedence.
func (s *Set) ResolveActionInStack(modeStack []Mode, combo string) (Binding, bool) {
	if s == nil {
		return Binding{}, false
	}
	norm := normalizeCombo(combo)
	if norm == "" {
		return Binding{}, false
	}

	for _, mode := range modeStack {
		entries := s.byMode[mode]
		if len(entries) == 0 {
			continue
		}
		b, ok := entries[norm]
		if !ok {
			continue
		}
		if b.Unbind {
			return Binding{}, false
		}
		if strings.TrimSpace(b.Action) == "" {
			continue
		}
		return b, true
	}

	return Binding{}, false
}

// ResolveWithGlobalFallback resolves a key with active modes plus global mode.
func (r *Resolver) ResolveWithGlobalFallback(activeModes []Mode, combo string) ChordResolveResult {
	return r.Resolve(ModeStack(activeModes, true), combo)
}
