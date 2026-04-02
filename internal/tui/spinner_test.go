package tui

import "testing"

func TestAllieSignatureSpinnerFrameSequence(t *testing.T) {
	want := []string{
		"[A....]",
		"[.A...]",
		"[..A..]",
		"[...A.]",
		"[....A]",
		"[...A.]",
		"[..A..]",
		"[.A...]",
	}

	got := allieSignatureSpinner.Frames
	if len(got) != len(want) {
		t.Fatalf("unexpected frame count: got %d want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("frame %d mismatch: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestNewSpinnerUsesAllieSignatureSpinnerByDefault(t *testing.T) {
	m := NewSpinner("thinking")
	if len(m.spinner.Spinner.Frames) != len(allieSignatureSpinner.Frames) {
		t.Fatalf("default spinner frame count mismatch: got %d want %d", len(m.spinner.Spinner.Frames), len(allieSignatureSpinner.Frames))
	}

	for i := range allieSignatureSpinner.Frames {
		if m.spinner.Spinner.Frames[i] != allieSignatureSpinner.Frames[i] {
			t.Fatalf("default spinner frame %d mismatch: got %q want %q", i, m.spinner.Spinner.Frames[i], allieSignatureSpinner.Frames[i])
		}
	}
}
