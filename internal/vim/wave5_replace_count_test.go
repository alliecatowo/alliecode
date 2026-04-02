package vim

import "testing"

func TestWave5ReplaceRespectsCountPrefix(t *testing.T) {
	e := NewEngine([]string{"abcdef"})
	e.HandleKeys("2lr")
	state := e.HandleKeys("Z")
	if state.Lines[0] != "abZdef" {
		t.Fatalf("expected replace to overwrite rune under cursor, got %#v", state.Lines)
	}
}
