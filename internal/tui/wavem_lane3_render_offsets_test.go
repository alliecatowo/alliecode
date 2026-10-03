package tui

import "testing"

func TestWaveMLane3_RenderModelOffsetsMonotonic(t *testing.T) {
	rows := []timelineEntry{
		{kind: timelineUser, text: "alpha", turn: 1},
		{kind: timelineAssistant, text: "beta beta beta beta beta", turn: 1},
		{kind: timelineAssistant, text: "gamma", turn: 1},
	}
	model := buildTimelineRenderModel(rows, "", 12)
	for i := 1; i < len(model.rows); i++ {
		if model.rows[i].lineOffset <= model.rows[i-1].lineOffset {
			t.Fatalf("expected increasing offsets, row %d offset=%d prev=%d", i, model.rows[i].lineOffset, model.rows[i-1].lineOffset)
		}
	}
}
