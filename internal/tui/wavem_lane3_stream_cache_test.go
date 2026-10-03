package tui

import "testing"

func TestWaveMLane3_StreamRenderCacheReuseAndInvalidate(t *testing.T) {
	app := New(Config{})
	app.streamBuf.WriteString("alpha beta gamma")
	blockA, linesA := app.streamRenderBlock(80)
	if blockA == "" || linesA == 0 {
		t.Fatalf("expected stream block for non-empty buffer")
	}
	blockB, linesB := app.streamRenderBlock(80)
	if blockA != blockB || linesA != linesB {
		t.Fatalf("expected cache hit for same width and content")
	}
	_, linesNarrow := app.streamRenderBlock(20)
	if linesNarrow <= 0 {
		t.Fatalf("expected recomputed lines for new width")
	}
	app.streamBuf.WriteString(" delta")
	blockC, _ := app.streamRenderBlock(20)
	if blockC == blockB {
		t.Fatalf("expected cache invalidation when stream grows")
	}
}
