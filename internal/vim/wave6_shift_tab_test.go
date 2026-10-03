package vim

import "testing"

func TestParseKeysShiftTabToken(t *testing.T) {
	keys := ParseKeys("<s-tab>")
	if len(keys) != 1 || keys[0].Special != SpecialShiftTab {
		t.Fatalf("expected shift+tab token parse, got %#v", keys)
	}
}

func TestHandleKeyWithTransitionLeavesModeStableOnShiftTab(t *testing.T) {
	e := NewEngine([]string{"alpha"})
	step := e.HandleKeyWithTransition(Key{Special: SpecialShiftTab})
	if step.Before.Mode != ModeNormal || step.After.Mode != ModeNormal {
		t.Fatalf("expected shift+tab to avoid vim mode churn, got before=%v after=%v", step.Before.Mode, step.After.Mode)
	}
}

func TestHandleKeyWithTransitionLeavesCursorStableOnShiftTabAndTab(t *testing.T) {
	e := NewEngine([]string{"alpha"})
	start := e.State().Cursor

	shiftTab := e.HandleKeyWithTransition(Key{Special: SpecialShiftTab})
	if shiftTab.After.Cursor != start {
		t.Fatalf("expected shift+tab cursor stable, got before=%+v after=%+v", start, shiftTab.After.Cursor)
	}

	tab := e.HandleKeyWithTransition(Key{Special: SpecialTab})
	if tab.After.Cursor != start {
		t.Fatalf("expected tab cursor stable, got before=%+v after=%+v", start, tab.After.Cursor)
	}
}
