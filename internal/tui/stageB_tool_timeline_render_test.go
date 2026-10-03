package tui

import (
	"strings"
	"testing"
)

func TestStageBToolTimelineRendersInputCardAndSummary(t *testing.T) {
	row := timelineEntry{
		kind:             timelineTool,
		turn:             3,
		toolName:         "bash",
		toolState:        toolProgressRunning,
		toolSummary:      "command=go test ./... | description=Run tests",
		toolInputPreview: "go test ./... PASS",
		toolInputBytes:   128,
	}
	rendered := renderTimelineRow(row, "", 72)
	for _, token := range []string{"TOOL bash [RUNNING]", "input: command=go test", "output: 128 chars", "summary:"} {
		if !strings.Contains(rendered, token) {
			t.Fatalf("expected %q in rendered row, got %q", token, rendered)
		}
	}
}

func TestStageBPermissionTimelineRendersRequestLine(t *testing.T) {
	row := timelineEntry{
		kind:            timelinePermission,
		turn:            5,
		toolName:        "bash",
		toolUseID:       "toolu_123",
		permissionState: permissionPending,
	}
	rendered := renderTimelineRow(row, "", 72)
	if !strings.Contains(rendered, "request: toolu_123") {
		t.Fatalf("expected request line in permission render, got %q", rendered)
	}
}
