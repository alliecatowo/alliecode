package tui

func resolveScrollAnchor(prevYOffset int, prevAtBottom, followTail bool, prevTotalLines, totalLines, prevViewportHeight, viewportHeight int) (int, bool) {
	if prevTotalLines < 0 {
		prevTotalLines = 0
	}
	if totalLines < 0 {
		totalLines = 0
	}
	if prevViewportHeight <= 0 {
		prevViewportHeight = 1
	}
	if viewportHeight <= 0 {
		viewportHeight = 1
	}
	if totalLines <= viewportHeight {
		return 0, false
	}
	if prevTotalLines <= 0 {
		if followTail || prevAtBottom {
			return 0, true
		}
		return clampYOffset(prevYOffset, totalLines, viewportHeight), false
	}
	if prevTotalLines <= prevViewportHeight && prevYOffset == 0 {
		if followTail || prevAtBottom || totalLines > prevTotalLines {
			return 0, true
		}
		return 0, false
	}

	prevMax := prevTotalLines - prevViewportHeight
	if prevMax < 0 {
		prevMax = 0
	}
	atBottomByOffset := prevMax > 0 && prevYOffset >= prevMax-1 && prevYOffset <= prevMax+1
	if followTail || prevAtBottom || atBottomByOffset {
		return 0, true
	}

	anchorOffset := prevYOffset
	if anchorOffset < 0 {
		anchorOffset = 0
	}
	prevAnchor := anchorOffset
	if prevAnchor > prevMax {
		prevAnchor = prevMax
	}

	if prevTotalLines == totalLines && viewportHeight != prevViewportHeight {
		return anchorOffset, false
	}

	if totalLines > prevTotalLines && viewportHeight != prevViewportHeight && prevTotalLines > prevViewportHeight {
		bottomGap := prevMax - prevAnchor
		if bottomGap < 0 {
			bottomGap = 0
		}
		return clampYOffset(totalLines-viewportHeight-bottomGap, totalLines, viewportHeight), false
	}

	if totalLines < prevTotalLines {
		return clampYOffset(anchorOffset, totalLines, viewportHeight), false
	}

	return clampYOffset(anchorOffset, totalLines, viewportHeight), false
}

func clampYOffset(offset, totalLines, viewportHeight int) int {
	if viewportHeight <= 0 {
		viewportHeight = 1
	}
	if totalLines <= viewportHeight {
		return 0
	}
	max := totalLines - viewportHeight
	if offset < 0 {
		return 0
	}
	if offset > max {
		return max
	}
	return offset
}
