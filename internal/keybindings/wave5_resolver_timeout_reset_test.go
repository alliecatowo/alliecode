package keybindings

import (
	"testing"
	"time"
)

func TestWave5ResolverResetSequenceClearsPendingChord(t *testing.T) {
	set := NewSet([]Binding{{Combo: "ctrl+k ctrl+o", Action: "open", Modes: []Mode{ModeGlobal}}})
	r := NewResolver(set, time.Second)
	started := r.Resolve([]Mode{ModeGlobal}, "ctrl+k")
	if started.Type != ResolveChordStarted {
		t.Fatalf("expected chord to start, got %+v", started)
	}
	r.ResetSequence()
	if len(r.Sequence().Pending) != 0 {
		t.Fatalf("expected pending sequence cleared")
	}
}
