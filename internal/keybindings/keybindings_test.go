package keybindings

import (
	"testing"
	"time"
)

func TestLookupAndLookupAll(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "Ctrl+K", Action: "open_palette", Modes: []Mode{ModeNormal, ModeInsert}},
		{Combo: "Ctrl+V", Action: "toggle_voice", Modes: []Mode{ModeNormal}},
	})

	b, ok := set.Lookup(ModeNormal, " ctrl+k ")
	if !ok || b.Action != "open_palette" {
		t.Fatalf("expected normal ctrl+k binding, got ok=%v binding=%+v", ok, b)
	}

	if _, ok := set.Lookup(ModeVoice, "ctrl+k"); ok {
		t.Fatal("did not expect voice mode ctrl+k")
	}

	all := set.LookupAll("CTRL+K")
	if len(all) != 1 || all[0].Action != "open_palette" {
		t.Fatalf("expected one ctrl+k binding, got %+v", all)
	}
}

func TestModeToggles(t *testing.T) {
	if got := ToggleMode(ModeNormal, ModeVoice); got != ModeVoice {
		t.Fatalf("expected toggle to target mode, got %q", got)
	}
	if got := ToggleMode(ModeVoice, ModeVoice); got != ModeNormal {
		t.Fatalf("expected toggle from target back to normal, got %q", got)
	}

	if got := NextMode(ModeNormal); got != ModeInsert {
		t.Fatalf("expected next mode insert, got %q", got)
	}
	if got := NextMode(ModeInsert); got != ModeVoice {
		t.Fatalf("expected next mode voice, got %q", got)
	}
	if got := NextMode(ModeVoice); got != ModeNormal {
		t.Fatalf("expected next mode normal, got %q", got)
	}
}

func TestLookupInModesLastMatchWins(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+k", Action: "normal_action", Modes: []Mode{ModeNormal}},
		{Combo: "ctrl+k", Action: "insert_action", Modes: []Mode{ModeInsert}},
	})

	b, ok := set.LookupInModes([]Mode{ModeNormal, ModeInsert}, "CTRL+K")
	if !ok {
		t.Fatal("expected ctrl+k to resolve in active contexts")
	}
	if b.Action != "insert_action" {
		t.Fatalf("expected last active binding to win, got %q", b.Action)
	}
}

func TestResolveWithChordInModes(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+k", Action: "single", Modes: []Mode{ModeNormal}},
		{Combo: "ctrl+k ctrl+s", Action: "save_all", Modes: []Mode{ModeNormal}},
		{Combo: "ctrl+k ctrl+d", Action: "dismiss", Modes: []Mode{ModeInsert}},
	})

	started := set.ResolveWithChordInModes([]Mode{ModeNormal}, nil, "ctrl+k")
	if started.Type != ResolveChordStarted || len(started.Pending) != 1 || started.Pending[0] != "ctrl+k" {
		t.Fatalf("expected chord start on ctrl+k prefix, got %+v", started)
	}

	matched := set.ResolveWithChordInModes([]Mode{ModeNormal}, started.Pending, "ctrl+s")
	if matched.Type != ResolveMatch || matched.Action != "save_all" {
		t.Fatalf("expected chord match save_all, got %+v", matched)
	}

	cancelled := set.ResolveWithChordInModes([]Mode{ModeNormal}, started.Pending, "x")
	if cancelled.Type != ResolveChordCancel {
		t.Fatalf("expected chord cancel for invalid continuation, got %+v", cancelled)
	}

	none := set.ResolveWithChordInModes([]Mode{ModeNormal}, nil, "ctrl+d")
	if none.Type != ResolveNone {
		t.Fatalf("expected ctrl+d to be inactive outside insert mode, got %+v", none)
	}
}

func TestResolveWithChordInModesRespectsUnbind(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+k ctrl+s", Action: "save_all", Modes: []Mode{ModeNormal}},
		{Combo: "ctrl+k ctrl+s", Action: "", Modes: []Mode{ModeNormal}, Unbind: true},
	})

	started := set.ResolveWithChordInModes([]Mode{ModeNormal}, nil, "ctrl+k")
	if started.Type != ResolveNone {
		t.Fatalf("expected no pending chord when only overridden by unbind, got %+v", started)
	}

	unbound := set.ResolveWithChordInModes([]Mode{ModeNormal}, []string{"ctrl+k"}, "ctrl+s")
	if unbound.Type != ResolveUnbound {
		t.Fatalf("expected unbound exact match, got %+v", unbound)
	}
}

func TestHintsForModeAndShortcutsForAction(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+o", Action: "search:quick_open", Modes: []Mode{ModeGlobal}},
		{Combo: "ctrl+shift+p", Action: "search:quick_open", Modes: []Mode{ModeGlobal}},
		{Combo: "ctrl+r", Action: "search:history", Modes: []Mode{ModeGlobal, ModeSearch}},
	})

	hints := set.HintsForMode(ModeGlobal, 2)
	if len(hints) != 2 {
		t.Fatalf("expected two hints with limit 2, got %d", len(hints))
	}

	shortcuts := set.ShortcutsForAction("search:quick_open")
	if len(shortcuts) != 2 {
		t.Fatalf("expected two shortcuts for action, got %+v", shortcuts)
	}
}

func TestHintsForModesDedupesAcrossLayers(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+o", Action: "search:quick_open", Modes: []Mode{ModeGlobal, ModeSearch}},
		{Combo: "ctrl+r", Action: "search:history", Modes: []Mode{ModeSearch}},
	})

	hints := set.HintsForModes([]Mode{ModeSearch, ModeGlobal}, 4)
	if len(hints) < 2 {
		t.Fatalf("expected aggregated hints, got %#v", hints)
	}
	if hints[0] != "ctrl+o -> search:quick_open" {
		t.Fatalf("expected deterministic first hint, got %#v", hints)
	}
}

func TestResolverMaintainsAndExpiresChordSequence(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+k ctrl+s", Action: "save_all", Modes: []Mode{ModeNormal}},
	})
	resolver := NewResolver(set, 20*time.Millisecond)
	now := time.Now()
	resolver.clock = func() time.Time { return now }

	started := resolver.Resolve([]Mode{ModeNormal}, "ctrl+k")
	if started.Type != ResolveChordStarted {
		t.Fatalf("expected chord started, got %+v", started)
	}
	if len(resolver.Sequence().Pending) != 1 {
		t.Fatalf("expected pending sequence to be stored")
	}

	now = now.Add(30 * time.Millisecond)
	matched := resolver.Resolve([]Mode{ModeNormal}, "ctrl+s")
	if matched.Type != ResolveNone {
		t.Fatalf("expected sequence timeout to clear pending state, got %+v", matched)
	}
}

func TestNormalizeComboCanonicalizesModifiersAndAliases(t *testing.T) {
	set := NewSet([]Binding{{Combo: "Cmd+Shift+P", Action: "palette", Modes: []Mode{ModeGlobal}}})
	b, ok := set.Lookup(ModeGlobal, "control+shift+p")
	if !ok || b.Action != "palette" {
		t.Fatalf("expected canonical combo lookup to resolve aliases, got ok=%v binding=%+v", ok, b)
	}
}
