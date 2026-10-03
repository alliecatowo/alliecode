package tui

import "testing"

func TestTimelineRenderPipelineSnapshotSimple(t *testing.T) {
	rows := []timelineEntry{{kind: timelineUser, text: "hello", turn: 1}, {kind: timelineAssistant, text: "done", turn: 1}}
	content, _, _, _, _ := renderTimeline(rows, "", 80)
	const want = "[01] USER\n  hello\n\n[01] AI\n  done"
	if content != want {
		t.Fatalf("snapshot mismatch\n--- got ---\n%s\n--- want ---\n%s", content, want)
	}
}
