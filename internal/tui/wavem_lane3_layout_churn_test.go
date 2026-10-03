package tui

import "testing"

func TestWaveMLane3_RecalcLayoutChurnKeepsViewportHeightPositive(t *testing.T) {
	app := readySizedApp(t, 120, 20)
	for i := 0; i < 30; i++ {
		app.width = 100 + (i % 3 * 8)
		app.height = 16 + (i % 4)
		app.recalcLayout()
		if app.viewport.Height < 1 {
			t.Fatalf("expected viewport height >= 1 during churn, got %d", app.viewport.Height)
		}
	}
}
