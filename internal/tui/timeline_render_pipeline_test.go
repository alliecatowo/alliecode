package tui

import "testing"

func TestBuildTimelineRenderModelFiltersAndOffsets(t *testing.T) {
	rows := []timelineEntry{{kind: timelineUser, text: "hello", turn: 1}, {kind: timelineAssistant, text: "world", turn: 1}, {kind: timelineTool, toolName: "bash", toolState: toolProgressRunning, turn: 1}}
	model := buildTimelineRenderModel(rows, "bash", 80)
	if len(model.rows) != 1 {
		t.Fatalf("expected one rendered row, got %d", len(model.rows))
	}
	if model.rows[0].rowIndex != 2 || model.rows[0].lineOffset != 0 {
		t.Fatalf("unexpected rendered row metadata: %#v", model.rows[0])
	}
}

func TestFlattenTimelineRenderModelProducesVisibleIndices(t *testing.T) {
	model := timelineRenderModel{rows: []timelineRenderedRow{{rowIndex: 3, block: "a", lineOffset: 0}, {rowIndex: 4, block: "b", lineOffset: 3}}}
	_, visible, offsets, _, _ := flattenTimelineRenderModel(model, 80)
	if len(visible) != 2 || visible[0] != 3 || visible[1] != 4 {
		t.Fatalf("unexpected visible rows: %#v", visible)
	}
	if len(offsets) != 2 || offsets[1] != 3 {
		t.Fatalf("unexpected offsets: %#v", offsets)
	}
}
