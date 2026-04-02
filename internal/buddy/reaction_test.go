package buddy

import "testing"

func TestAnalyzeReaction_DirectNameAndKeywords(t *testing.T) {
	companion := Companion{
		CompanionSoul: CompanionSoul{Name: "Mochi"},
		CompanionBones: CompanionBones{Stats: map[StatName]int{
			StatSnark: 10,
		}},
	}

	got, ok := AnalyzeReaction(companion, "Mochi thanks for the save")
	if !ok {
		t.Fatalf("expected reaction detected")
	}
	if got != "tiny salute" {
		t.Fatalf("unexpected reaction: %q", got)
	}
}

func TestAnalyzeReaction_DefaultByStats(t *testing.T) {
	companion := Companion{
		CompanionSoul: CompanionSoul{Name: "Mochi"},
		CompanionBones: CompanionBones{Stats: map[StatName]int{
			StatSnark: 80,
		}},
	}

	got, ok := AnalyzeReaction(companion, "mochi check this")
	if !ok {
		t.Fatalf("expected reaction detected")
	}
	if got != "bold move" {
		t.Fatalf("unexpected reaction: %q", got)
	}
}

func TestAnalyzeReaction_NoTrigger(t *testing.T) {
	companion := Companion{CompanionSoul: CompanionSoul{Name: "Mochi"}}
	got, ok := AnalyzeReaction(companion, "just regular drafting update")
	if ok || got != "" {
		t.Fatalf("expected no reaction, got %q ok=%v", got, ok)
	}
}

func TestAnalyzeReaction_ContextualGlobalTriggers(t *testing.T) {
	companion := Companion{CompanionSoul: CompanionSoul{Name: "Mochi"}}

	cases := []struct {
		text string
		want string
	}{
		{text: "need help with this test", want: "you got this"},
		{text: "build failed with exception", want: "debug mode paws on"},
		{text: "all tests pass now", want: "nice fix"},
		{text: "optimized and faster now", want: "zoom zoom"},
		{text: "ship it", want: "lets go"},
	}

	for _, tc := range cases {
		got, ok := AnalyzeReaction(companion, tc.text)
		if !ok {
			t.Fatalf("expected trigger for %q", tc.text)
		}
		if got != tc.want {
			t.Fatalf("for %q expected %q got %q", tc.text, tc.want, got)
		}
	}
}

func TestAnalyzeReaction_ContextualDirectTriggers(t *testing.T) {
	companion := Companion{CompanionSoul: CompanionSoul{Name: "Mochi"}}

	cases := []struct {
		text string
		want string
	}{
		{text: "Mochi panic deploy", want: "breathing first, then fixes"},
		{text: "mochi got traceback", want: "errors are clues"},
		{text: "Mochi this is slow", want: "speed pass activated"},
		{text: "Mochi fixed it", want: "clean recovery"},
		{text: "Mochi we merged", want: "celebration wiggle"},
		{text: "Mochi sorry", want: "all good, onward"},
	}

	for _, tc := range cases {
		got, ok := AnalyzeReaction(companion, tc.text)
		if !ok {
			t.Fatalf("expected direct trigger for %q", tc.text)
		}
		if got != tc.want {
			t.Fatalf("for %q expected %q got %q", tc.text, tc.want, got)
		}
	}
}
