package vim

import "testing"

func TestNormalMotions(t *testing.T) {
	e := NewEngine([]string{"one two", "three"})

	e.HandleKeys("lll")
	if got := e.State().Cursor; got != (Cursor{Row: 0, Col: 3}) {
		t.Fatalf("expected cursor 0,3 got %+v", got)
	}

	e.HandleKeys("j")
	if got := e.State().Cursor; got != (Cursor{Row: 1, Col: 3}) {
		t.Fatalf("expected cursor 1,3 got %+v", got)
	}

	e.HandleKeys("0")
	if got := e.State().Cursor; got != (Cursor{Row: 1, Col: 0}) {
		t.Fatalf("expected cursor 1,0 got %+v", got)
	}

	e.HandleKeys("$")
	if got := e.State().Cursor; got != (Cursor{Row: 1, Col: 4}) {
		t.Fatalf("expected cursor 1,4 got %+v", got)
	}

	e.HandleKeys("k")
	if got := e.State().Cursor; got != (Cursor{Row: 0, Col: 4}) {
		t.Fatalf("expected cursor 0,4 got %+v", got)
	}

	e.HandleKeys("h")
	if got := e.State().Cursor; got != (Cursor{Row: 0, Col: 3}) {
		t.Fatalf("expected cursor 0,3 got %+v", got)
	}
}

func TestWordMotions(t *testing.T) {
	e := NewEngine([]string{"alpha beta gamma"})

	e.HandleKeys("w")
	if got := e.State().Cursor; got != (Cursor{Row: 0, Col: 6}) {
		t.Fatalf("expected w to move to beta start, got %+v", got)
	}

	e.HandleKeys("e")
	if got := e.State().Cursor; got != (Cursor{Row: 0, Col: 9}) {
		t.Fatalf("expected e to move to beta end, got %+v", got)
	}

	e.HandleKeys("b")
	if got := e.State().Cursor; got != (Cursor{Row: 0, Col: 6}) {
		t.Fatalf("expected b to move to beta start, got %+v", got)
	}
}

func TestInsertAppendOpenAndEscape(t *testing.T) {
	e := NewEngine([]string{"cat"})

	e.HandleKeys("i")
	if e.State().Mode != ModeInsert {
		t.Fatalf("expected insert mode")
	}
	e.HandleKeys("X")
	e.HandleKey(Key{Special: SpecialEsc})

	st := e.State()
	if st.Lines[0] != "Xcat" {
		t.Fatalf("unexpected line: %q", st.Lines[0])
	}
	if st.Mode != ModeNormal {
		t.Fatalf("expected normal mode after esc")
	}

	e = NewEngine([]string{"cat"})
	e.HandleKeys("a")
	e.HandleKeys("s")
	e.HandleKey(Key{Special: SpecialEsc})
	if got := e.State().Lines[0]; got != "csat" {
		t.Fatalf("append expected csat, got %q", got)
	}

	e = NewEngine([]string{"cat"})
	e.HandleKeys("o")
	if e.State().Mode != ModeInsert {
		t.Fatalf("expected insert mode for o")
	}
	e.HandleKeys("dog")
	e.HandleKey(Key{Special: SpecialEsc})
	if got := e.State().Lines; len(got) != 2 || got[1] != "dog" {
		t.Fatalf("unexpected lines after o: %#v", got)
	}
}

func TestDeleteOperations(t *testing.T) {
	e := NewEngine([]string{"alpha beta", "second"})

	e.HandleKeys("x")
	if got := e.State().Lines[0]; got != "lpha beta" {
		t.Fatalf("x delete failed: %q", got)
	}

	e = NewEngine([]string{"alpha beta"})
	e.HandleKeys("dw")
	if got := e.State().Lines[0]; got != "beta" {
		t.Fatalf("dw failed: %q", got)
	}

	e = NewEngine([]string{"alpha beta"})
	e.HandleKeys("diw")
	if got := e.State().Lines[0]; got != " beta" {
		t.Fatalf("diw failed: %q", got)
	}

	e = NewEngine([]string{"first", "second", "third"})
	e.HandleKeys("dd")
	if got := e.State().Lines; len(got) != 2 || got[0] != "second" {
		t.Fatalf("dd failed: %#v", got)
	}
}

func TestYankPasteAndUndo(t *testing.T) {
	e := NewEngine([]string{"one", "two", "three"})

	e.HandleKeys("yy")
	e.HandleKeys("p")
	st := e.State()
	if len(st.Lines) != 4 || st.Lines[1] != "one" {
		t.Fatalf("yy+p failed: %#v", st.Lines)
	}

	e.HandleKeys("u")
	st = e.State()
	if len(st.Lines) != 3 || st.Lines[1] != "two" {
		t.Fatalf("undo failed: %#v", st.Lines)
	}

	e = NewEngine([]string{"alpha beta"})
	e.HandleKeys("yw")
	e.HandleKeys("p")
	if got := e.State().Lines[0]; got != "aalpha lpha beta" {
		t.Fatalf("charwise yank paste failed: %q", got)
	}
}

func TestOperatorPendingTransitions(t *testing.T) {
	e := NewEngine([]string{"alpha"})
	e.HandleKeys("d")
	st := e.State()
	if st.Mode != ModeOperatorPending || st.PendingOperator != OperatorDelete {
		t.Fatalf("expected operator pending delete, got mode=%v op=%v", st.Mode, st.PendingOperator)
	}

	e.HandleKey(Key{Special: SpecialEsc})
	st = e.State()
	if st.Mode != ModeNormal || st.PendingOperator != OperatorNone {
		t.Fatalf("expected operator clear on esc, got mode=%v op=%v", st.Mode, st.PendingOperator)
	}
}

func TestCountsApplyToMotionsAndOperators(t *testing.T) {
	e := NewEngine([]string{"one two three four", "five six"})

	e.HandleKeys("3w")
	if got := e.State().Cursor; got != (Cursor{Row: 0, Col: 14}) {
		t.Fatalf("expected 3w to reach four, got %+v", got)
	}

	e.HandleKeys("2b")
	if got := e.State().Cursor; got != (Cursor{Row: 0, Col: 4}) {
		t.Fatalf("expected 2b to reach two, got %+v", got)
	}

	e = NewEngine([]string{"one two three four"})
	e.HandleKeys("2dw")
	if got := e.State().Lines[0]; got != "three four" {
		t.Fatalf("expected 2dw to delete two words, got %q", got)
	}

	e = NewEngine([]string{"one two three four"})
	e.HandleKeys("d2w")
	if got := e.State().Lines[0]; got != "three four" {
		t.Fatalf("expected d2w to delete two words, got %q", got)
	}

	e = NewEngine([]string{"a", "b", "c", "d"})
	e.HandleKeys("2dd")
	if got := e.State().Lines; len(got) != 2 || got[0] != "c" || got[1] != "d" {
		t.Fatalf("expected 2dd to delete two lines, got %#v", got)
	}
}

func TestAdditionalTextObjectsAndChangeOperator(t *testing.T) {
	e := NewEngine([]string{"alpha  beta"})
	e.HandleKeys("daw")
	if got := e.State().Lines[0]; got != "beta" {
		t.Fatalf("expected daw to remove word and spaces, got %q", got)
	}

	e = NewEngine([]string{"alpha beta gamma"})
	e.HandleKeys("2daw")
	if got := e.State().Lines[0]; got != "gamma" {
		t.Fatalf("expected 2daw to remove two words, got %q", got)
	}

	e = NewEngine([]string{"alpha beta"})
	e.HandleKeys("ciw")
	if mode := e.State().Mode; mode != ModeInsert {
		t.Fatalf("expected ciw to enter insert mode, got %v", mode)
	}
	e.HandleKeys("ONE")
	e.HandleKey(Key{Special: SpecialEsc})
	if got := e.State().Lines[0]; got != "ONE beta" {
		t.Fatalf("expected ciw replacement, got %q", got)
	}

	e = NewEngine([]string{"one two three"})
	e.HandleKeys("2cw")
	if mode := e.State().Mode; mode != ModeInsert {
		t.Fatalf("expected 2cw to enter insert mode, got %v", mode)
	}
	e.HandleKeys("X")
	e.HandleKey(Key{Special: SpecialEsc})
	if got := e.State().Lines[0]; got != "Xthree" {
		t.Fatalf("expected 2cw to change two words, got %q", got)
	}
}

func TestGAndGgMotionsWithCounts(t *testing.T) {
	e := NewEngine([]string{"one", "two", "three", "four"})

	e.HandleKeys("G")
	if got := e.State().Cursor; got != (Cursor{Row: 3, Col: 0}) {
		t.Fatalf("expected G to move to last line start, got %+v", got)
	}

	e.HandleKeys("gg")
	if got := e.State().Cursor; got != (Cursor{Row: 0, Col: 0}) {
		t.Fatalf("expected gg to move to first line start, got %+v", got)
	}

	e.HandleKeys("3gg")
	if got := e.State().Cursor; got != (Cursor{Row: 2, Col: 0}) {
		t.Fatalf("expected 3gg to move to line 3, got %+v", got)
	}

	e.HandleKeys("2G")
	if got := e.State().Cursor; got != (Cursor{Row: 1, Col: 0}) {
		t.Fatalf("expected 2G to move to line 2, got %+v", got)
	}
}

func TestOperatorGAndGgLinewiseRanges(t *testing.T) {
	e := NewEngine([]string{"one", "two", "three", "four"})

	e.HandleKeys("j")
	e.HandleKeys("dG")
	if got := e.State().Lines; len(got) != 1 || got[0] != "one" {
		t.Fatalf("expected dG to delete through last line, got %#v", got)
	}

	e = NewEngine([]string{"one", "two", "three", "four"})
	e.HandleKeys("G")
	e.HandleKeys("dgg")
	if got := e.State().Lines; len(got) != 1 || got[0] != "" {
		t.Fatalf("expected dgg from last line to delete full buffer, got %#v", got)
	}

	e = NewEngine([]string{"one", "two", "three", "four"})
	e.HandleKeys("j")
	e.HandleKeys("cG")
	if mode := e.State().Mode; mode != ModeInsert {
		t.Fatalf("expected cG to enter insert mode, got %v", mode)
	}
	e.HandleKeys("X")
	e.HandleKey(Key{Special: SpecialEsc})
	if got := e.State().Lines; len(got) != 1 || got[0] != "Xone" {
		t.Fatalf("expected cG replacement from current line to end, got %#v", got)
	}
}

func TestInsertModeSpecialKeysAndEditing(t *testing.T) {
	e := NewEngine([]string{"alpha", "beta"})
	e.HandleKeys("i")
	e.HandleKey(Key{Special: SpecialEnd})
	e.HandleKey(Key{Special: SpecialEnter})
	e.HandleKeys("X")
	e.HandleKey(Key{Special: SpecialBackspace})
	e.HandleKeys("Y")
	e.HandleKey(Key{Special: SpecialUp})
	e.HandleKey(Key{Special: SpecialHome})
	e.HandleKeys("Z")
	e.HandleKey(Key{Special: SpecialEsc})

	st := e.State()
	if len(st.Lines) != 3 {
		t.Fatalf("expected newline split in insert mode, got %#v", st.Lines)
	}
	if st.Lines[0] != "Zalpha" {
		t.Fatalf("expected home/up insertion result, got %q", st.Lines[0])
	}
	if st.Lines[1] != "Y" {
		t.Fatalf("expected enter/backspace edit result, got %q", st.Lines[1])
	}
}

func TestParserSpecialTokensAndNormalSpecialMotions(t *testing.T) {
	keys := ParseKeys("<left>")
	if len(keys) != 1 || keys[0].Special != SpecialLeft {
		t.Fatalf("expected special left token parse, got %#v", keys)
	}

	e := NewEngine([]string{"abc", "def"})
	e.HandleKeys("ll")
	e.HandleKey(Key{Special: SpecialDown})
	e.HandleKey(Key{Special: SpecialEnd})
	if got := e.State().Cursor; got != (Cursor{Row: 1, Col: 2}) {
		t.Fatalf("expected normal-mode special cursor movement, got %+v", got)
	}
}

func TestUppercaseInsertVariantsAndCaretMotion(t *testing.T) {
	e := NewEngine([]string{"  alpha"})
	e.HandleKeys("^")
	if got := e.State().Cursor; got != (Cursor{Row: 0, Col: 2}) {
		t.Fatalf("expected caret motion to first non-space, got %+v", got)
	}

	e.HandleKeys("A")
	e.HandleKeys("!")
	e.HandleKey(Key{Special: SpecialEsc})
	if got := e.State().Lines[0]; got != "  alpha!" {
		t.Fatalf("expected A append at end, got %q", got)
	}

	e = NewEngine([]string{"  alpha"})
	e.HandleKeys("I")
	e.HandleKeys("X")
	e.HandleKey(Key{Special: SpecialEsc})
	if got := e.State().Lines[0]; got != "  Xalpha" {
		t.Fatalf("expected I insert at first non-space, got %q", got)
	}

	e = NewEngine([]string{"one", "two"})
	e.HandleKeys("j")
	e.HandleKeys("O")
	e.HandleKeys("top")
	e.HandleKey(Key{Special: SpecialEsc})
	if got := e.State().Lines; len(got) != 3 || got[1] != "top" {
		t.Fatalf("expected O to open above current line, got %#v", got)
	}
}

func TestReplaceSubstituteAndDeleteToEndVariants(t *testing.T) {
	e := NewEngine([]string{"alpha beta"})
	e.HandleKeys("rX")
	if got := e.State().Lines[0]; got != "Xlpha beta" {
		t.Fatalf("expected r to replace character, got %q", got)
	}

	e = NewEngine([]string{"alpha beta"})
	e.HandleKeys("s")
	if mode := e.State().Mode; mode != ModeInsert {
		t.Fatalf("expected s to enter insert mode, got %v", mode)
	}
	e.HandleKeys("Z")
	e.HandleKey(Key{Special: SpecialEsc})
	if got := e.State().Lines[0]; got != "Zlpha beta" {
		t.Fatalf("expected s substitution result, got %q", got)
	}

	e = NewEngine([]string{"alpha beta"})
	e.HandleKeys("D")
	if got := e.State().Lines[0]; got != "" {
		t.Fatalf("expected D to delete to end of line, got %q", got)
	}

	e = NewEngine([]string{"alpha beta"})
	e.HandleKeys("C")
	if mode := e.State().Mode; mode != ModeInsert {
		t.Fatalf("expected C to enter insert mode, got %v", mode)
	}
}

func TestPasteBeforeWithUppercaseP(t *testing.T) {
	e := NewEngine([]string{"one", "two"})
	e.HandleKeys("yy")
	e.HandleKeys("j")
	e.HandleKeys("P")
	lines := e.State().Lines
	if len(lines) != 3 || lines[1] != "one" {
		t.Fatalf("expected P to paste line before cursor line, got %#v", lines)
	}
}
