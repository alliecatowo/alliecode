package keybindings

import (
	"sort"
	"strings"
	"time"
)

type Mode string

const (
	ModeGlobal       Mode = "global"
	ModeChat         Mode = "chat"
	ModeSearch       Mode = "search"
	ModeAutocomplete Mode = "autocomplete"
	ModePermission   Mode = "permission"
	ModeVim          Mode = "vim"
	ModeSlash        Mode = "slash"
	ModeReference    Mode = "reference"
	ModeQuickOpen    Mode = "quick_open"
	ModeHistory      Mode = "history_search"
	ModeTimeline     Mode = "timeline_search"
	ModeModelPicker  Mode = "model_picker"
	ModeCommandPanel Mode = "command_panel"

	ModeNormal Mode = "normal"
	ModeInsert Mode = "insert"
	ModeVoice  Mode = "voice"
)

type Binding struct {
	Combo  string
	Action string
	Modes  []Mode
	Unbind bool
}

type Set struct {
	bindings []Binding
	byMode   map[Mode]map[string]Binding
	chords   map[Mode]map[string]struct{}
}

type ResolveType string

const (
	ResolveMatch        ResolveType = "match"
	ResolveNone         ResolveType = "none"
	ResolveUnbound      ResolveType = "unbound"
	ResolveChordStarted ResolveType = "chord_started"
	ResolveChordCancel  ResolveType = "chord_cancelled"
)

type ChordResolveResult struct {
	Type    ResolveType
	Action  string
	Pending []string
}

type SequenceState struct {
	Pending      []string
	PendingSince time.Time
}

type Resolver struct {
	set      *Set
	clock    func() time.Time
	timeout  time.Duration
	sequence SequenceState
}

type Shortcut struct {
	Mode   Mode
	Combo  string
	Action string
}

func NewSet(bindings []Binding) *Set {
	byMode := map[Mode]map[string]Binding{}
	chords := map[Mode]map[string]struct{}{}
	normalized := make([]Binding, 0, len(bindings))
	for _, b := range bindings {
		combo := normalizeCombo(b.Combo)
		if combo == "" {
			continue
		}
		copyBinding := b
		copyBinding.Combo = combo
		normalized = append(normalized, copyBinding)
		for _, mode := range copyBinding.Modes {
			if byMode[mode] == nil {
				byMode[mode] = map[string]Binding{}
			}
			if chords[mode] == nil {
				chords[mode] = map[string]struct{}{}
			}
			byMode[mode][combo] = copyBinding
			parts := splitChord(combo)
			if len(parts) > 1 {
				for i := 1; i < len(parts); i++ {
					prefix := strings.Join(parts[:i], " ")
					chords[mode][prefix] = struct{}{}
				}
			}
		}
	}
	return &Set{bindings: normalized, byMode: byMode, chords: chords}
}

func NewResolver(set *Set, timeout time.Duration) *Resolver {
	if timeout <= 0 {
		timeout = 1250 * time.Millisecond
	}
	return &Resolver{set: set, timeout: timeout, clock: time.Now}
}

func (r *Resolver) Sequence() SequenceState {
	out := SequenceState{PendingSince: r.sequence.PendingSince}
	if len(r.sequence.Pending) > 0 {
		out.Pending = append([]string(nil), r.sequence.Pending...)
	}
	return out
}

func (r *Resolver) ResetSequence() {
	r.sequence = SequenceState{}
}

func (r *Resolver) Resolve(modes []Mode, combo string) ChordResolveResult {
	if r == nil || r.set == nil {
		return ChordResolveResult{Type: ResolveNone}
	}
	now := time.Now()
	if r.clock != nil {
		now = r.clock()
	}
	if len(r.sequence.Pending) > 0 && !r.sequence.PendingSince.IsZero() && now.Sub(r.sequence.PendingSince) > r.timeout {
		r.sequence = SequenceState{}
	}
	result := r.set.ResolveWithChordInModes(modes, r.sequence.Pending, combo)
	switch result.Type {
	case ResolveChordStarted:
		r.sequence.Pending = append([]string(nil), result.Pending...)
		r.sequence.PendingSince = now
	case ResolveMatch, ResolveUnbound, ResolveChordCancel:
		r.sequence = SequenceState{}
	}
	return result
}

func (s *Set) Lookup(mode Mode, combo string) (Binding, bool) {
	if s == nil {
		return Binding{}, false
	}
	entries := s.byMode[mode]
	if entries == nil {
		return Binding{}, false
	}
	b, ok := entries[normalizeCombo(combo)]
	return b, ok
}

func (s *Set) LookupInModes(modes []Mode, combo string) (Binding, bool) {
	if s == nil {
		return Binding{}, false
	}
	norm := normalizeCombo(combo)
	if norm == "" {
		return Binding{}, false
	}
	var out Binding
	found := false
	for _, mode := range modes {
		entries := s.byMode[mode]
		if entries == nil {
			continue
		}
		if b, ok := entries[norm]; ok {
			out = b
			found = true
		}
	}
	return out, found
}

func (s *Set) LookupAll(combo string) []Binding {
	if s == nil {
		return nil
	}
	norm := normalizeCombo(combo)
	out := make([]Binding, 0)
	for _, b := range s.bindings {
		if b.Combo == norm {
			out = append(out, b)
		}
	}
	return out
}

func ToggleMode(current, target Mode) Mode {
	if current == target {
		return ModeNormal
	}
	return target
}

func NextMode(current Mode) Mode {
	switch current {
	case ModeNormal:
		return ModeInsert
	case ModeInsert:
		return ModeVoice
	default:
		return ModeNormal
	}
}

func normalizeCombo(combo string) string {
	norm := strings.ToLower(strings.TrimSpace(combo))
	if norm == "" {
		return ""
	}
	steps := strings.Fields(norm)
	for i, step := range steps {
		step = normalizeTerminalSequenceAlias(step)
		parts := strings.Split(step, "+")
		ctrl := false
		alt := false
		shift := false
		key := ""
		for _, part := range parts {
			part = normalizeTerminalSequenceAlias(strings.TrimSpace(part))
			switch part {
			case "ctrl", "control", "cmd", "command":
				ctrl = true
			case "alt", "option", "meta":
				alt = true
			case "shift":
				shift = true
			case "s-tab", "backtab", "iso-left-tab":
				shift = true
				key = "tab"
			case "s-enter":
				shift = true
				key = "enter"
			case "esc":
				key = "escape"
			case "return":
				key = "enter"
			case "pgup":
				key = "pageup"
			case "pgdn", "pgdown", "next":
				key = "pagedown"
			case "prior":
				key = "pageup"
			default:
				if part != "" {
					key = part
				}
			}
		}
		ordered := make([]string, 0, 4)
		if ctrl {
			ordered = append(ordered, "ctrl")
		}
		if alt {
			ordered = append(ordered, "alt")
		}
		if shift {
			ordered = append(ordered, "shift")
		}
		if key != "" {
			ordered = append(ordered, key)
		}
		steps[i] = strings.Join(ordered, "+")
	}
	return strings.Join(steps, " ")
}

func normalizeTerminalSequenceAlias(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "\x1b", "")
	if s == "" {
		return ""
	}
	switch s {
	case "[z", "[1;2z", "[9;2u", "[27;2;9~", "[z", "1;2z", "9;2u", "27;2;9~", "back-tab", "kcbt", "btab":
		return "shift+tab"
	case "[13;2u", "[27;2;13~", "13;2u", "27;2;13~":
		return "shift+enter"
	default:
		return s
	}
}

func splitChord(combo string) []string {
	norm := normalizeCombo(combo)
	if norm == "" {
		return nil
	}
	parts := strings.Fields(norm)
	if len(parts) == 0 {
		return nil
	}
	return parts
}

func prefixMatches(prefix []string, full []string) bool {
	if len(prefix) >= len(full) {
		return false
	}
	for i := range prefix {
		if prefix[i] != full[i] {
			return false
		}
	}
	return true
}

func exactMatches(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (s *Set) ResolveWithChordInModes(modes []Mode, pending []string, combo string) ChordResolveResult {
	if s == nil {
		return ChordResolveResult{Type: ResolveNone}
	}
	stroke := normalizeCombo(combo)
	if stroke == "" {
		if len(pending) > 0 {
			return ChordResolveResult{Type: ResolveChordCancel}
		}
		return ChordResolveResult{Type: ResolveNone}
	}

	testChord := make([]string, 0, len(pending)+1)
	testChord = append(testChord, pending...)
	testChord = append(testChord, stroke)

	modeSet := map[Mode]bool{}
	for _, m := range modes {
		modeSet[m] = true
	}

	type chordWinner struct {
		action  string
		unbind  bool
		present bool
	}

	chordWinners := map[string]chordWinner{}

	hasLongerChord := false
	for _, b := range s.bindings {
		if len(b.Modes) == 0 {
			continue
		}
		active := false
		for _, m := range b.Modes {
			if modeSet[m] {
				active = true
				break
			}
		}
		if !active {
			continue
		}
		parts := splitChord(b.Combo)
		if len(parts) == 0 {
			continue
		}
		chordKey := strings.Join(parts, " ")
		chordWinners[chordKey] = chordWinner{action: b.Action, unbind: b.Unbind, present: true}
	}

	for _, b := range s.bindings {
		parts := splitChord(b.Combo)
		if len(parts) == 0 {
			continue
		}
		if !prefixMatches(testChord, parts) {
			continue
		}
		winner, ok := chordWinners[strings.Join(parts, " ")]
		if ok && winner.present && !winner.unbind {
			hasLongerChord = true
		}
	}
	if hasLongerChord {
		pendingOut := make([]string, len(testChord))
		copy(pendingOut, testChord)
		return ChordResolveResult{Type: ResolveChordStarted, Pending: pendingOut}
	}

	var exact *Binding
	for i := range s.bindings {
		b := &s.bindings[i]
		if len(b.Modes) == 0 {
			continue
		}
		active := false
		for _, m := range b.Modes {
			if modeSet[m] {
				active = true
				break
			}
		}
		if !active {
			continue
		}
		parts := splitChord(b.Combo)
		if len(parts) == 0 {
			continue
		}
		if exactMatches(testChord, parts) {
			exact = b
		}
	}

	if exact != nil {
		if exact.Unbind {
			return ChordResolveResult{Type: ResolveUnbound}
		}
		return ChordResolveResult{Type: ResolveMatch, Action: exact.Action}
	}

	if len(pending) > 0 {
		return ChordResolveResult{Type: ResolveChordCancel}
	}
	return ChordResolveResult{Type: ResolveNone}
}

func (s *Set) ActionsForMode(mode Mode) []string {
	if s == nil {
		return nil
	}
	entries := s.byMode[mode]
	if len(entries) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(entries))
	for _, b := range entries {
		action := strings.TrimSpace(b.Action)
		if action == "" || b.Unbind {
			continue
		}
		if _, ok := seen[action]; ok {
			continue
		}
		seen[action] = struct{}{}
		out = append(out, action)
	}
	sort.Strings(out)
	return out
}

func (s *Set) HintsForMode(mode Mode, limit int) []string {
	if s == nil {
		return nil
	}
	if limit <= 0 {
		limit = 6
	}
	entries := s.byMode[mode]
	if len(entries) == 0 {
		return nil
	}
	keys := make([]string, 0, len(entries))
	for combo := range entries {
		keys = append(keys, combo)
	}
	sort.Slice(keys, func(i, j int) bool {
		iChord := strings.Contains(keys[i], " ")
		jChord := strings.Contains(keys[j], " ")
		if iChord != jChord {
			return !iChord
		}
		return keys[i] < keys[j]
	})
	out := make([]string, 0, min(limit, len(keys)))
	for _, combo := range keys {
		if len(out) >= limit {
			break
		}
		b := entries[combo]
		if b.Unbind {
			continue
		}
		action := strings.TrimSpace(b.Action)
		if action == "" {
			continue
		}
		out = append(out, combo+" -> "+action)
	}
	return out
}

func (s *Set) HintsForModes(modes []Mode, limit int) []string {
	if s == nil || len(modes) == 0 {
		return nil
	}
	if limit <= 0 {
		limit = 6
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, limit)
	for _, mode := range modes {
		for _, hint := range s.HintsForMode(mode, limit) {
			if len(out) >= limit {
				return out
			}
			if _, ok := seen[hint]; ok {
				continue
			}
			seen[hint] = struct{}{}
			out = append(out, hint)
		}
	}
	return out
}

func (s *Set) ShortcutsForAction(action string) []Shortcut {
	if s == nil {
		return nil
	}
	action = strings.ToLower(strings.TrimSpace(action))
	if action == "" {
		return nil
	}
	out := make([]Shortcut, 0, 4)
	for mode, entries := range s.byMode {
		for combo, b := range entries {
			if b.Unbind {
				continue
			}
			if strings.ToLower(strings.TrimSpace(b.Action)) != action {
				continue
			}
			out = append(out, Shortcut{Mode: mode, Combo: combo, Action: b.Action})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Mode != out[j].Mode {
			return out[i].Mode < out[j].Mode
		}
		return out[i].Combo < out[j].Combo
	})
	return out
}

func (s *Set) HasChordPrefixInModes(modes []Mode, prefix []string) bool {
	if s == nil || len(prefix) == 0 {
		return false
	}
	p := strings.Join(prefix, " ")
	for _, mode := range modes {
		if entries := s.chords[mode]; entries != nil {
			if _, ok := entries[p]; ok {
				return true
			}
		}
	}
	return false
}
