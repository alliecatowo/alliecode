package vim

import "testing"

func TestOpenAboveInsertsBlankLine(t *testing.T) {
	b := newTextBuffer([]string{"one", "two"})
	c := b.openAbove(1)
	if c.Row != 1 || c.Col != 0 {
		t.Fatalf("expected cursor at inserted line, got %+v", c)
	}
	if len(b.lines) != 3 || b.lines[1] != "" {
		t.Fatalf("expected blank inserted line, got %#v", b.lines)
	}
}
