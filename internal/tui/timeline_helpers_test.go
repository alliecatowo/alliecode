package tui

import (
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/permissions"
)

func TestBuildToolPreviewTruncates(t *testing.T) {
	out := buildToolPreview("abc", "defghi", 8)
	if out != "abcde..." {
		t.Fatalf("expected truncated preview, got %q", out)
	}
}

func TestBuildToolPreviewNoTruncate(t *testing.T) {
	out := buildToolPreview("abc", " def", 32)
	if out != "abc def" {
		t.Fatalf("expected merged preview, got %q", out)
	}
}

func TestNextMatchPosWraps(t *testing.T) {
	if got := nextMatchPos(2, 3, 1); got != 0 {
		t.Fatalf("expected forward wrap to 0, got %d", got)
	}
	if got := nextMatchPos(0, 3, -1); got != 2 {
		t.Fatalf("expected backward wrap to 2, got %d", got)
	}
}

func TestNextMatchPosHandlesOutOfRangeCurrent(t *testing.T) {
	if got := nextMatchPos(-1, 3, 1); got != 0 {
		t.Fatalf("expected invalid low index to snap to first, got %d", got)
	}
	if got := nextMatchPos(9, 3, -1); got != 2 {
		t.Fatalf("expected invalid high index to snap to last on reverse, got %d", got)
	}
}

func TestPermissionModeLabel(t *testing.T) {
	if permissionModeLabel(permissions.ModePlan) != "plan" {
		t.Fatalf("expected plan label")
	}
	if permissionModeLabel(permissions.ModeBypass) != "bypass" {
		t.Fatalf("expected bypass label")
	}
}

func TestRenderTimelineFiltersRows(t *testing.T) {
	rows := []timelineEntry{
		{kind: timelineUser, text: "hello world", turn: 1},
		{kind: timelineTool, toolName: "bash", toolState: toolProgressRunning, turn: 1},
		{kind: timelineAssistant, text: "done", turn: 1},
	}

	content, visible, offsets, total := renderTimeline(rows, "bash", 80)
	if len(visible) != 1 || visible[0] != 1 {
		t.Fatalf("expected only tool row visible, got %#v", visible)
	}
	if len(offsets) != 1 || offsets[0] != 0 {
		t.Fatalf("expected first visible row offset 0, got %#v", offsets)
	}
	if total != 1 {
		t.Fatalf("expected one visual line, got %d", total)
	}
	if !strings.Contains(strings.ToLower(content), "tool bash") {
		t.Fatalf("expected rendered content to include tool row, got %q", content)
	}
}

func TestRenderTimelineOffsetsRespectWrappedWidths(t *testing.T) {
	rows := []timelineEntry{
		{kind: timelineUser, text: "12345678901234567890", turn: 1},
		{kind: timelineTool, toolName: "bash", toolState: toolProgressRunning, turn: 1},
	}

	_, _, offsets, _ := renderTimeline(rows, "", 10)
	if len(offsets) != 2 {
		t.Fatalf("expected two rows, got %v", offsets)
	}
	if offsets[1] != 6 {
		t.Fatalf("expected second offset 6, got %d", offsets[1])
	}
}

func TestRenderTimelinePermissionStatuses(t *testing.T) {
	rows := []timelineEntry{
		{kind: timelinePermission, toolName: "bash", permissionState: permissionPending, turn: 1},
		{kind: timelinePermission, toolName: "bash", permissionState: permissionApproved, turn: 1},
		{kind: timelinePermission, toolName: "bash", permissionState: permissionDenied, turn: 1},
		{kind: timelinePermission, toolName: "bash", permissionState: permissionAlwaysStatus, turn: 1},
	}

	content, _, _, _ := renderTimeline(rows, "", 120)
	lower := strings.ToLower(content)
	for _, token := range []string{"[pending]", "[approved]", "[denied]", "[always]"} {
		if !strings.Contains(lower, token) {
			t.Fatalf("expected permission status token %q in render output, got %q", token, content)
		}
	}
}
