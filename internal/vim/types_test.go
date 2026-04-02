package vim

import "testing"

func TestStateCarriesLinesSnapshot(t *testing.T) {
	e := NewEngine([]string{"one"})
	st := e.State()
	if len(st.Lines) != 1 || st.Lines[0] != "one" {
		t.Fatalf("unexpected state lines snapshot: %#v", st.Lines)
	}
}
