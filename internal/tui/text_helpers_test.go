package tui

import (
	"fmt"
	"strings"
	"testing"
)

func TestRuneSafeBackspaceUnicode(t *testing.T) {
	s := "go🙂"
	got := runeSafeBackspace(s)
	if got != "go" {
		t.Fatalf("expected rune-safe delete, got %q", got)
	}
}

func TestTruncateDisplayWidthWideRune(t *testing.T) {
	got := truncateDisplayWidth("ab界cd", 5, "...")
	if got != "ab..." {
		t.Fatalf("expected width-safe truncation, got %q", got)
	}
}

func TestVisualLineCountWrap(t *testing.T) {
	got := visualLineCount("1234567890", 4)
	if got != 3 {
		t.Fatalf("expected 3 wrapped lines, got %d", got)
	}
}

func TestHighlightMatchesCaseInsensitive(t *testing.T) {
	out := highlightMatches("Hello tools", "TOO")
	if !strings.Contains(strings.ToLower(stripANSI(out)), "hello tools") {
		t.Fatalf("expected base text preserved, got %q", out)
	}
	if got := highlightMatches("Hello tools", "xyz"); got != "Hello tools" {
		t.Fatalf("expected no-op for missing query, got %q", got)
	}
}

func TestHighlightMatchesUnicodeSafe(t *testing.T) {
	out := highlightMatches("Cafe cafe CAFE", "cafe")
	plain := stripANSI(out)
	if plain != "Cafe cafe CAFE" {
		t.Fatalf("expected text preserved after highlighting, got %q", plain)
	}
}

func TestHighlightMatchesEscapesRegexMeta(t *testing.T) {
	out := highlightMatches("a+b a?b [ab]", "a+b")
	plain := stripANSI(out)
	if plain != "a+b a?b [ab]" {
		t.Fatalf("expected literal query matching, got %q", plain)
	}
}

func TestHighlightMatchesOverlappingRangesMerge(t *testing.T) {
	out := highlightMatches("ababa", "aba")
	if plain := stripANSI(out); plain != "ababa" {
		t.Fatalf("expected text preservation, got %q", plain)
	}
	highlight := searchHighlightStyle.Render("ababa")
	if strings.Count(out, highlight) != 1 {
		t.Fatalf("expected one merged highlight, got %q", out)
	}
}

func TestHighlightMatchesAdjacentRangesMerge(t *testing.T) {
	out := highlightMatches("aaaa", "aa")
	highlight := searchHighlightStyle.Render("aaaa")
	if strings.Count(out, highlight) != 1 {
		t.Fatalf("expected adjacent overlap to merge into single span, got %q", out)
	}
}

func TestOverlapMatchRangesInsensitive(t *testing.T) {
	ranges := overlapMatchRangesInsensitive("BaNaNa", "ana")
	if len(ranges) != 1 || ranges[0][0] != 1 || ranges[0][1] != 6 {
		t.Fatalf("expected merged overlap [1,6), got %v", ranges)
	}

	emptyCases := [][2]string{{"", "ana"}, {"banana", ""}, {"short", "muchlonger"}}
	for _, c := range emptyCases {
		if got := overlapMatchRangesInsensitive(c[0], c[1]); len(got) != 0 {
			t.Fatalf("expected no ranges for %q/%q, got %v", c[0], c[1], got)
		}
	}

	out := highlightMatches("banana", "ana")
	merged := searchHighlightStyle.Render("anana")
	if !strings.Contains(out, merged) {
		t.Fatalf("expected merged styled segment %q in %q", fmt.Sprintf("%q", merged), out)
	}
}
