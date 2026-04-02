package vim

import "strings"

type rangeSelection struct {
	start Cursor
	end   Cursor
}

type textBuffer struct {
	lines []string
}

func newTextBuffer(lines []string) *textBuffer {
	if len(lines) == 0 {
		return &textBuffer{lines: []string{""}}
	}
	out := make([]string, len(lines))
	copy(out, lines)
	return &textBuffer{lines: out}
}

func (b *textBuffer) cloneLines() []string {
	out := make([]string, len(b.lines))
	copy(out, b.lines)
	return out
}

func (b *textBuffer) lineCount() int {
	return len(b.lines)
}

func (b *textBuffer) line(row int) string {
	if row < 0 || row >= len(b.lines) {
		return ""
	}
	return b.lines[row]
}

func (b *textBuffer) lineLen(row int) int {
	return len(b.line(row))
}

func clamp(n, min, max int) int {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

func (b *textBuffer) clampCursor(c Cursor) Cursor {
	if len(b.lines) == 0 {
		b.lines = []string{""}
	}
	row := clamp(c.Row, 0, len(b.lines)-1)
	col := clamp(c.Col, 0, len(b.lines[row]))
	return Cursor{Row: row, Col: col}
}

func (b *textBuffer) insertRune(c Cursor, r rune) Cursor {
	c = b.clampCursor(c)
	line := b.lines[c.Row]
	b.lines[c.Row] = line[:c.Col] + string(r) + line[c.Col:]
	return Cursor{Row: c.Row, Col: c.Col + 1}
}

func (b *textBuffer) insertNewline(c Cursor) Cursor {
	c = b.clampCursor(c)
	line := b.lines[c.Row]
	left := line[:c.Col]
	right := line[c.Col:]
	b.lines[c.Row] = left
	b.lines = append(b.lines[:c.Row+1], append([]string{right}, b.lines[c.Row+1:]...)...)
	return Cursor{Row: c.Row + 1, Col: 0}
}

func (b *textBuffer) openBelow(row int) Cursor {
	row = clamp(row, 0, len(b.lines)-1)
	b.lines = append(b.lines[:row+1], append([]string{""}, b.lines[row+1:]...)...)
	return Cursor{Row: row + 1, Col: 0}
}

func (b *textBuffer) openAbove(row int) Cursor {
	row = clamp(row, 0, len(b.lines)-1)
	b.lines = append(b.lines[:row], append([]string{""}, b.lines[row:]...)...)
	return Cursor{Row: row, Col: 0}
}

func (b *textBuffer) deleteRange(sel rangeSelection) string {
	start := b.clampCursor(sel.start)
	end := b.clampCursor(sel.end)
	if start.Row > end.Row || (start.Row == end.Row && start.Col > end.Col) {
		start, end = end, start
	}
	if start.Row == end.Row && start.Col == end.Col {
		return ""
	}

	if start.Row == end.Row {
		line := b.lines[start.Row]
		deleted := line[start.Col:end.Col]
		b.lines[start.Row] = line[:start.Col] + line[end.Col:]
		return deleted
	}

	first := b.lines[start.Row]
	last := b.lines[end.Row]
	deletedLines := []string{first[start.Col:]}
	for row := start.Row + 1; row < end.Row; row++ {
		deletedLines = append(deletedLines, b.lines[row])
	}
	deletedLines = append(deletedLines, last[:end.Col])

	b.lines[start.Row] = first[:start.Col] + last[end.Col:]
	b.lines = append(b.lines[:start.Row+1], b.lines[end.Row+1:]...)
	if len(b.lines) == 0 {
		b.lines = []string{""}
	}

	return strings.Join(deletedLines, "\n")
}

func (b *textBuffer) deleteLine(row int) string {
	if len(b.lines) == 0 {
		b.lines = []string{""}
		return ""
	}
	row = clamp(row, 0, len(b.lines)-1)
	deleted := b.lines[row]
	b.lines = append(b.lines[:row], b.lines[row+1:]...)
	if len(b.lines) == 0 {
		b.lines = []string{""}
	}
	return deleted
}

func (b *textBuffer) insertAfter(c Cursor, text string, linewise bool) Cursor {
	c = b.clampCursor(c)
	if linewise {
		insertAt := c.Row + 1
		parts := strings.Split(text, "\n")
		if len(parts) == 0 {
			parts = []string{""}
		}
		for i := len(parts) - 1; i >= 0; i-- {
			b.lines = append(b.lines[:insertAt], append([]string{parts[i]}, b.lines[insertAt:]...)...)
		}
		return Cursor{Row: insertAt, Col: 0}
	}

	line := b.lines[c.Row]
	insertCol := c.Col + 1
	if insertCol > len(line) {
		insertCol = len(line)
	}

	parts := strings.Split(text, "\n")
	if len(parts) == 1 {
		b.lines[c.Row] = line[:insertCol] + parts[0] + line[insertCol:]
		return Cursor{Row: c.Row, Col: insertCol}
	}

	first := line[:insertCol] + parts[0]
	last := parts[len(parts)-1] + line[insertCol:]
	middle := parts[1 : len(parts)-1]
	newLines := append([]string{first}, middle...)
	newLines = append(newLines, last)
	b.lines = append(b.lines[:c.Row], append(newLines, b.lines[c.Row+1:]...)...)
	return Cursor{Row: c.Row, Col: insertCol}
}

func (b *textBuffer) insertBefore(c Cursor, text string, linewise bool) Cursor {
	c = b.clampCursor(c)
	if linewise {
		insertAt := c.Row
		parts := strings.Split(text, "\n")
		if len(parts) == 0 {
			parts = []string{""}
		}
		for i := len(parts) - 1; i >= 0; i-- {
			b.lines = append(b.lines[:insertAt], append([]string{parts[i]}, b.lines[insertAt:]...)...)
		}
		return Cursor{Row: insertAt, Col: 0}
	}

	line := b.lines[c.Row]
	insertCol := c.Col
	if insertCol > len(line) {
		insertCol = len(line)
	}

	parts := strings.Split(text, "\n")
	if len(parts) == 1 {
		b.lines[c.Row] = line[:insertCol] + parts[0] + line[insertCol:]
		return Cursor{Row: c.Row, Col: insertCol}
	}

	first := line[:insertCol] + parts[0]
	last := parts[len(parts)-1] + line[insertCol:]
	middle := parts[1 : len(parts)-1]
	newLines := append([]string{first}, middle...)
	newLines = append(newLines, last)
	b.lines = append(b.lines[:c.Row], append(newLines, b.lines[c.Row+1:]...)...)
	return Cursor{Row: c.Row, Col: insertCol}
}

func (b *textBuffer) deleteBefore(c Cursor) Cursor {
	c = b.clampCursor(c)
	if c.Row == 0 && c.Col == 0 {
		return c
	}
	if c.Col > 0 {
		line := b.lines[c.Row]
		b.lines[c.Row] = line[:c.Col-1] + line[c.Col:]
		return Cursor{Row: c.Row, Col: c.Col - 1}
	}
	prevRow := c.Row - 1
	prev := b.lines[prevRow]
	cur := b.lines[c.Row]
	b.lines[prevRow] = prev + cur
	b.lines = append(b.lines[:c.Row], b.lines[c.Row+1:]...)
	return Cursor{Row: prevRow, Col: len(prev)}
}

func (b *textBuffer) deleteAt(c Cursor) Cursor {
	c = b.clampCursor(c)
	line := b.lines[c.Row]
	if c.Col < len(line) {
		b.lines[c.Row] = line[:c.Col] + line[c.Col+1:]
		return b.clampCursor(c)
	}
	if c.Row+1 >= len(b.lines) {
		return c
	}
	b.lines[c.Row] = line + b.lines[c.Row+1]
	b.lines = append(b.lines[:c.Row+1], b.lines[c.Row+2:]...)
	return b.clampCursor(c)
}
