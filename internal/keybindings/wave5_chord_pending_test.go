package keybindings

import "testing"

func TestWave5ResolverSequenceSnapshotIsCopy(t *testing.T) {
	set := NewSet([]Binding{{Combo: "ctrl+k ctrl+o", Action: "open", Modes: []Mode{ModeGlobal}}})
	r := NewResolver(set, 0)
	_ = r.Resolve([]Mode{ModeGlobal}, "ctrl+k")
	snap := r.Sequence()
	if len(snap.Pending) == 0 {
		t.Fatalf("expected pending sequence")
	}
	snap.Pending[0] = "mutated"
	if r.Sequence().Pending[0] == "mutated" {
		t.Fatalf("expected resolver sequence snapshot to be immutable copy")
	}
}
