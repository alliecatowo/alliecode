package keybindings

import "testing"

func TestModeStackDedupAndGlobalAppend(t *testing.T) {
	stack := ModeStack([]Mode{ModeSearch, ModeSearch, ModeAutocomplete}, true)
	if len(stack) != 3 {
		t.Fatalf("expected deduped stack with global fallback, got %#v", stack)
	}
	if stack[0] != ModeSearch || stack[1] != ModeAutocomplete || stack[2] != ModeGlobal {
		t.Fatalf("unexpected stack ordering: %#v", stack)
	}
}

func TestResolveActionInStackHonorsPrecedenceAndUnbind(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+o", Action: "global-open", Modes: []Mode{ModeGlobal}},
		{Combo: "ctrl+o", Action: "search-open", Modes: []Mode{ModeSearch}},
		{Combo: "ctrl+x", Action: "global-x", Modes: []Mode{ModeGlobal}},
		{Combo: "ctrl+x", Action: "", Modes: []Mode{ModeSearch}, Unbind: true},
	})

	b, ok := set.ResolveActionInStack([]Mode{ModeSearch, ModeGlobal}, "ctrl+o")
	if !ok || b.Action != "search-open" {
		t.Fatalf("expected search layer to win, got ok=%t binding=%+v", ok, b)
	}

	if _, ok := set.ResolveActionInStack([]Mode{ModeSearch, ModeGlobal}, "ctrl+x"); ok {
		t.Fatalf("expected search unbind to block global fallback")
	}
}

func TestResolverResolveWithGlobalFallback(t *testing.T) {
	set := NewSet([]Binding{
		{Combo: "ctrl+f", Action: "search:timeline", Modes: []Mode{ModeGlobal}},
	})
	r := NewResolver(set, 0)

	result := r.ResolveWithGlobalFallback([]Mode{ModeChat}, "ctrl+f")
	if result.Type != ResolveMatch || result.Action != "search:timeline" {
		t.Fatalf("expected global fallback match, got %+v", result)
	}
}

func TestModeStackWithoutGlobal(t *testing.T) {
	stack := ModeStack([]Mode{ModeInsert, ModeChat, ModeInsert}, false)
	if len(stack) != 2 {
		t.Fatalf("expected deduped stack without global, got %#v", stack)
	}
	if stack[0] != ModeInsert || stack[1] != ModeChat {
		t.Fatalf("unexpected stack ordering: %#v", stack)
	}
}
