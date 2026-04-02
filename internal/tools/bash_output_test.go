package tools

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestFormatCommandTimeout(t *testing.T) {
	out := formatCommandTimeout(2*time.Second, "partial output")
	if !strings.Contains(out, "timed out") {
		t.Fatalf("expected timeout text, got: %q", out)
	}
	if !strings.Contains(out, "partial output") {
		t.Fatalf("expected output body, got: %q", out)
	}
}

func TestFormatCommandExit(t *testing.T) {
	err := errors.New("exit status 7")
	out := formatCommandExit("oops", err)
	if !strings.Contains(out, "oops") || !strings.Contains(out, "Exit code: exit status 7") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestTruncateCommandOutputByLinesAndChars(t *testing.T) {
	out := truncateCommandOutput("a\nb\nc", 2, 100)
	want := "a\nb\n... (truncated to 2 lines)"
	if out != want {
		t.Fatalf("unexpected line truncation\nwant: %q\n got: %q", want, out)
	}

	out = truncateCommandOutput("1234567890", 0, 5)
	if out != "12345\n... (truncated)" {
		t.Fatalf("unexpected char truncation: %q", out)
	}
}
