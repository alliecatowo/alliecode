package tui

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStageBToolInputSummaryBash(t *testing.T) {
	raw := json.RawMessage(`{"command":"go test ./...","description":"Run full tests"}`)
	summary := summarizeToolInput("bash", raw)
	if !strings.Contains(summary, "command=go test ./...") {
		t.Fatalf("expected command in summary, got %q", summary)
	}
	if !strings.Contains(summary, "description=Run full tests") {
		t.Fatalf("expected description in summary, got %q", summary)
	}
}

func TestStageBToolInputSummaryUnknownFallsBackToJSON(t *testing.T) {
	raw := json.RawMessage(`{"foo":"bar","count":2}`)
	summary := summarizeToolInput("custom_tool", raw)
	if !strings.Contains(summary, "foo") || !strings.Contains(summary, "bar") {
		t.Fatalf("expected compact json fallback summary, got %q", summary)
	}
}
