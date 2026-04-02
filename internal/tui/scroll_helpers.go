package tui

func resolveScrollAnchor(prevYOffset int, prevAtBottom, followTail bool, prevTotalLines, totalLines, viewportHeight int) (int, bool) {
	if viewportHeight <= 0 {
		viewportHeight = 1
	}
	if totalLines <= viewportHeight {
		return 0, false
	}
	if prevTotalLines <= viewportHeight && prevYOffset == 0 {
		return 0, true
	}
	if followTail || prevAtBottom {
		return 0, true
	}
	max := totalLines - viewportHeight
	if prevYOffset < 0 {
		return 0, false
	}
	if prevYOffset > max {
		return max, false
	}
	return prevYOffset, false
}
