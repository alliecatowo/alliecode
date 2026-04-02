package vim

import "unicode"

type snapshot struct {
	lines            []string
	cursor           Cursor
	mode             Mode
	pendingOperator  Operator
	register         string
	registerLinewise bool
}

// Engine is a deterministic Vim-like modal editing state machine.
type Engine struct {
	buf                *textBuffer
	cursor             Cursor
	mode               Mode
	pendingOperator    Operator
	awaitingTextObject rune
	awaitingG          bool
	pendingGCount      int
	awaitingOperatorG  bool
	pendingOperatorG   int
	countPrefix        int
	operatorCount      int
	awaitingReplace    bool
	replaceCount       int
	register           string
	registerLinewise   bool
	undoStack          []snapshot
}

func NewEngine(lines []string) *Engine {
	b := newTextBuffer(lines)
	e := &Engine{buf: b, mode: ModeNormal}
	e.cursor = e.normalizeNormalCursor(Cursor{})
	return e
}

func (e *Engine) State() State {
	return State{
		Mode:            e.mode,
		Cursor:          e.cursor,
		PendingOperator: e.pendingOperator,
		Lines:           e.buf.cloneLines(),
	}
}

func (e *Engine) HandleKeys(input string) State {
	keys := ParseKeys(input)
	for _, k := range keys {
		e.HandleKey(k)
	}
	return e.State()
}

func (e *Engine) HandleKey(k Key) State {
	if k.Special == SpecialEsc {
		e.handleEsc()
		return e.State()
	}

	if e.mode == ModeInsert {
		e.handleInsert(k)
		return e.State()
	}

	if k.Special != SpecialNone {
		e.handleNormalSpecial(k)
		return e.State()
	}

	r := k.Rune
	if e.mode == ModeOperatorPending {
		e.handleOperatorPending(r)
		return e.State()
	}

	e.handleNormal(r)
	return e.State()
}

func (e *Engine) handleNormalSpecial(k Key) {
	switch k.Special {
	case SpecialLeft:
		e.cursor = e.normalizeNormalCursor(Cursor{Row: e.cursor.Row, Col: e.cursor.Col - 1})
	case SpecialRight:
		e.cursor = e.normalizeNormalCursor(Cursor{Row: e.cursor.Row, Col: e.cursor.Col + 1})
	case SpecialUp:
		e.cursor = e.normalizeNormalCursor(Cursor{Row: e.cursor.Row - 1, Col: e.cursor.Col})
	case SpecialDown:
		e.cursor = e.normalizeNormalCursor(Cursor{Row: e.cursor.Row + 1, Col: e.cursor.Col})
	case SpecialHome:
		e.cursor = e.normalizeNormalCursor(Cursor{Row: e.cursor.Row, Col: 0})
	case SpecialEnd:
		lineLen := e.buf.lineLen(e.cursor.Row)
		if lineLen == 0 {
			e.cursor = Cursor{Row: e.cursor.Row, Col: 0}
		} else {
			e.cursor = Cursor{Row: e.cursor.Row, Col: lineLen - 1}
		}
	}
}

func (e *Engine) handleEsc() {
	e.pendingOperator = OperatorNone
	e.awaitingTextObject = 0
	e.awaitingG = false
	e.pendingGCount = 0
	e.awaitingOperatorG = false
	e.pendingOperatorG = 0
	e.countPrefix = 0
	e.operatorCount = 0
	e.awaitingReplace = false
	e.replaceCount = 0
	if e.mode == ModeInsert {
		e.mode = ModeNormal
		e.cursor = e.normalizeNormalCursor(Cursor{Row: e.cursor.Row, Col: e.cursor.Col - 1})
		return
	}
	e.mode = ModeNormal
}

func (e *Engine) handleInsert(k Key) {
	if k.Special != SpecialNone {
		switch k.Special {
		case SpecialEnter:
			e.pushUndo()
			e.cursor = e.buf.insertNewline(e.cursor)
		case SpecialBackspace:
			e.deleteBeforeCursor()
		case SpecialDelete:
			e.deleteUnderCursorInsertMode()
		case SpecialLeft:
			e.cursor = e.buf.clampCursor(Cursor{Row: e.cursor.Row, Col: e.cursor.Col - 1})
		case SpecialRight:
			e.cursor = e.buf.clampCursor(Cursor{Row: e.cursor.Row, Col: e.cursor.Col + 1})
		case SpecialUp:
			e.cursor = e.buf.clampCursor(Cursor{Row: e.cursor.Row - 1, Col: e.cursor.Col})
		case SpecialDown:
			e.cursor = e.buf.clampCursor(Cursor{Row: e.cursor.Row + 1, Col: e.cursor.Col})
		case SpecialHome:
			e.cursor = e.buf.clampCursor(Cursor{Row: e.cursor.Row, Col: 0})
		case SpecialEnd:
			e.cursor = e.buf.clampCursor(Cursor{Row: e.cursor.Row, Col: e.buf.lineLen(e.cursor.Row)})
		}
		return
	}
	if k.Rune == '\n' {
		e.pushUndo()
		e.cursor = e.buf.insertNewline(e.cursor)
		return
	}
	e.pushUndo()
	e.cursor = e.buf.insertRune(e.cursor, k.Rune)
}

func (e *Engine) deleteBeforeCursor() {
	if e.cursor.Row == 0 && e.cursor.Col == 0 {
		return
	}
	e.pushUndo()
	e.cursor = e.buf.deleteBefore(e.cursor)
}

func (e *Engine) deleteUnderCursorInsertMode() {
	e.pushUndo()
	e.cursor = e.buf.deleteAt(e.cursor)
}

func (e *Engine) handleNormal(r rune) {
	if e.awaitingReplace {
		count := e.replaceCount
		if count <= 0 {
			count = 1
		}
		e.awaitingReplace = false
		e.replaceCount = 0
		e.replaceUnderCursor(r, count)
		return
	}

	if e.awaitingG {
		count := e.pendingGCount
		if count <= 0 {
			count = 1
		}
		e.awaitingG = false
		e.pendingGCount = 0
		if r == 'g' {
			targetRow := 0
			if count > 1 {
				targetRow = count - 1
			}
			e.cursor = e.normalizeNormalCursor(Cursor{Row: targetRow, Col: 0})
		}
		return
	}

	if r >= '0' && r <= '9' {
		if r == '0' && e.countPrefix == 0 {
			e.cursor = e.normalizeNormalCursor(Cursor{Row: e.cursor.Row, Col: 0})
			return
		}
		e.countPrefix = e.countPrefix*10 + int(r-'0')
		return
	}

	count := e.consumeCount()
	switch r {
	case 'h':
		for i := 0; i < count; i++ {
			e.cursor = e.normalizeNormalCursor(Cursor{Row: e.cursor.Row, Col: e.cursor.Col - 1})
		}
	case 'j':
		for i := 0; i < count; i++ {
			e.cursor = e.normalizeNormalCursor(Cursor{Row: e.cursor.Row + 1, Col: e.cursor.Col})
		}
	case 'k':
		for i := 0; i < count; i++ {
			e.cursor = e.normalizeNormalCursor(Cursor{Row: e.cursor.Row - 1, Col: e.cursor.Col})
		}
	case 'l':
		for i := 0; i < count; i++ {
			e.cursor = e.normalizeNormalCursor(Cursor{Row: e.cursor.Row, Col: e.cursor.Col + 1})
		}
	case 'w':
		for i := 0; i < count; i++ {
			e.cursor = e.normalizeNormalCursor(e.motionWordForward(e.cursor))
		}
	case 'b':
		for i := 0; i < count; i++ {
			e.cursor = e.normalizeNormalCursor(e.motionWordBackward(e.cursor))
		}
	case 'e':
		for i := 0; i < count; i++ {
			e.cursor = e.normalizeNormalCursor(e.motionWordEnd(e.cursor))
		}
	case '$':
		row := clamp(e.cursor.Row+count-1, 0, e.buf.lineCount()-1)
		lineLen := e.buf.lineLen(row)
		if lineLen == 0 {
			e.cursor = Cursor{Row: row, Col: 0}
		} else {
			e.cursor = Cursor{Row: row, Col: lineLen - 1}
		}
	case '^':
		row := clamp(e.cursor.Row+count-1, 0, e.buf.lineCount()-1)
		line := e.buf.line(row)
		col := 0
		for col < len(line) && unicode.IsSpace(rune(line[col])) {
			col++
		}
		e.cursor = e.normalizeNormalCursor(Cursor{Row: row, Col: col})
	case 'g':
		e.awaitingG = true
		e.pendingGCount = count
	case 'G':
		targetRow := e.buf.lineCount() - 1
		if count > 1 {
			targetRow = count - 1
		}
		e.cursor = e.normalizeNormalCursor(Cursor{Row: targetRow, Col: 0})
	case 'i':
		e.mode = ModeInsert
		e.pendingOperator = OperatorNone
	case 'a':
		e.mode = ModeInsert
		e.pendingOperator = OperatorNone
		e.cursor = e.buf.clampCursor(Cursor{Row: e.cursor.Row, Col: e.cursor.Col + 1})
	case 'I':
		line := e.buf.line(e.cursor.Row)
		col := 0
		for col < len(line) && unicode.IsSpace(rune(line[col])) {
			col++
		}
		e.cursor = e.buf.clampCursor(Cursor{Row: e.cursor.Row, Col: col})
		e.mode = ModeInsert
		e.pendingOperator = OperatorNone
	case 'A':
		e.cursor = e.buf.clampCursor(Cursor{Row: e.cursor.Row, Col: e.buf.lineLen(e.cursor.Row)})
		e.mode = ModeInsert
		e.pendingOperator = OperatorNone
	case 'o':
		e.pushUndo()
		e.cursor = e.buf.openBelow(e.cursor.Row)
		e.mode = ModeInsert
		e.pendingOperator = OperatorNone
	case 'O':
		e.pushUndo()
		e.cursor = e.buf.openAbove(e.cursor.Row)
		e.mode = ModeInsert
		e.pendingOperator = OperatorNone
	case 'x':
		for i := 0; i < count; i++ {
			e.deleteCharUnderCursor()
		}
	case 's':
		for i := 0; i < count; i++ {
			e.deleteCharUnderCursor()
		}
		e.mode = ModeInsert
		e.pendingOperator = OperatorNone
	case 'S':
		e.applyLineOperator(e.cursor.Row, count)
		e.mode = ModeInsert
		e.pendingOperator = OperatorNone
	case 'C':
		if sel, ok := e.selectionForMotion('$', count); ok {
			e.pendingOperator = OperatorChange
			e.applyOperatorSelection(sel)
			e.pendingOperator = OperatorNone
			e.mode = ModeInsert
		}
	case 'D':
		if sel, ok := e.selectionForMotion('$', count); ok {
			e.pendingOperator = OperatorDelete
			e.applyOperatorSelection(sel)
			e.pendingOperator = OperatorNone
			e.mode = ModeNormal
		}
	case 'r':
		e.awaitingReplace = true
		e.replaceCount = count
	case 'd':
		e.mode = ModeOperatorPending
		e.pendingOperator = OperatorDelete
		e.awaitingTextObject = 0
		e.operatorCount = count
	case 'y':
		e.mode = ModeOperatorPending
		e.pendingOperator = OperatorYank
		e.awaitingTextObject = 0
		e.operatorCount = count
	case 'c':
		e.mode = ModeOperatorPending
		e.pendingOperator = OperatorChange
		e.awaitingTextObject = 0
		e.operatorCount = count
	case 'p':
		for i := 0; i < count; i++ {
			e.pasteAfter()
		}
	case 'P':
		for i := 0; i < count; i++ {
			e.pasteBefore()
		}
	case 'u':
		for i := 0; i < count; i++ {
			e.undo()
		}
	}
}

func (e *Engine) handleOperatorPending(r rune) {
	if e.awaitingOperatorG {
		op := e.pendingOperator
		if r == 'g' {
			targetRow := 0
			if e.pendingOperatorG > 1 {
				targetRow = e.pendingOperatorG - 1
			}
			e.applyOperatorToLineTarget(targetRow)
		}
		e.clearOperatorState()
		if op == OperatorChange {
			e.mode = ModeInsert
		} else {
			e.mode = ModeNormal
		}
		return
	}

	if r >= '0' && r <= '9' {
		e.countPrefix = e.countPrefix*10 + int(r-'0')
		return
	}

	motionCount := e.consumeCount()
	totalCount := e.operatorCount * motionCount
	if totalCount <= 0 {
		totalCount = 1
	}

	if e.awaitingTextObject != 0 {
		op := e.pendingOperator
		if r == 'w' {
			if e.awaitingTextObject == 'i' {
				sel := e.innerWordSelection(e.cursor, totalCount)
				e.applyOperatorSelection(sel)
			} else if e.awaitingTextObject == 'a' {
				sel := e.aroundWordSelection(e.cursor, totalCount)
				e.applyOperatorSelection(sel)
			}
		}
		e.clearOperatorState()
		if op == OperatorChange {
			e.mode = ModeInsert
		} else {
			e.mode = ModeNormal
		}
		return
	}

	if (e.pendingOperator == OperatorDelete && r == 'd') || (e.pendingOperator == OperatorYank && r == 'y') || (e.pendingOperator == OperatorChange && r == 'c') {
		op := e.pendingOperator
		e.applyLineOperator(e.cursor.Row, totalCount)
		e.clearOperatorState()
		if op == OperatorChange {
			e.mode = ModeInsert
		} else {
			e.mode = ModeNormal
		}
		return
	}

	if r == 'i' {
		e.awaitingTextObject = 'i'
		return
	}

	if r == 'a' {
		e.awaitingTextObject = 'a'
		return
	}

	if r == 'g' {
		e.awaitingOperatorG = true
		e.pendingOperatorG = totalCount
		return
	}

	if r == 'G' {
		targetRow := e.buf.lineCount() - 1
		if totalCount > 1 {
			targetRow = totalCount - 1
		}
		e.applyOperatorToLineTarget(targetRow)
		if e.pendingOperator == OperatorChange {
			e.mode = ModeInsert
		} else {
			e.mode = ModeNormal
		}
		e.clearOperatorState()
		return
	}

	if sel, ok := e.selectionForMotion(r, totalCount); ok {
		e.applyOperatorSelection(sel)
	}
	if e.pendingOperator == OperatorChange {
		e.mode = ModeInsert
	} else {
		e.mode = ModeNormal
	}
	e.clearOperatorState()
}

func (e *Engine) selectionForMotion(motion rune, count int) (rangeSelection, bool) {
	start := e.buf.clampCursor(Cursor{Row: e.cursor.Row, Col: e.cursor.Col})
	var end Cursor
	if count <= 0 {
		count = 1
	}
	switch motion {
	case 'h':
		end = e.buf.clampCursor(Cursor{Row: start.Row, Col: start.Col - count})
		if end.Row == start.Row && end.Col == start.Col {
			return rangeSelection{}, false
		}
		return rangeSelection{start: end, end: start}, true
	case 'l':
		end = e.buf.clampCursor(Cursor{Row: start.Row, Col: start.Col + count})
		if end.Row == start.Row && end.Col == start.Col {
			return rangeSelection{}, false
		}
		return rangeSelection{start: start, end: end}, true
	case 'w':
		end = start
		for i := 0; i < count; i++ {
			end = e.motionWordForwardExclusive(end)
		}
		if end.Row == start.Row && end.Col == start.Col {
			return rangeSelection{}, false
		}
		return rangeSelection{start: start, end: end}, true
	case 'b':
		end = start
		for i := 0; i < count; i++ {
			end = e.motionWordBackward(end)
		}
		if end.Row == start.Row && end.Col == start.Col {
			return rangeSelection{}, false
		}
		return rangeSelection{start: end, end: start}, true
	case 'e':
		end = start
		for i := 0; i < count; i++ {
			wEnd := e.motionWordEnd(end)
			end = e.buf.clampCursor(Cursor{Row: wEnd.Row, Col: wEnd.Col + 1})
		}
		if end.Row == start.Row && end.Col == start.Col {
			return rangeSelection{}, false
		}
		return rangeSelection{start: start, end: end}, true
	case '0':
		end = Cursor{Row: start.Row, Col: 0}
		if end.Col == start.Col {
			return rangeSelection{}, false
		}
		return rangeSelection{start: end, end: start}, true
	case '$':
		row := clamp(start.Row+count-1, 0, e.buf.lineCount()-1)
		end = Cursor{Row: row, Col: e.buf.lineLen(row)}
		if end.Col == start.Col {
			return rangeSelection{}, false
		}
		return rangeSelection{start: start, end: end}, true
	default:
		return rangeSelection{}, false
	}
}

func (e *Engine) applyLineOperator(row, count int) {
	row = clamp(row, 0, e.buf.lineCount()-1)
	if count <= 0 {
		count = 1
	}
	endRow := clamp(row+count-1, 0, e.buf.lineCount()-1)
	e.applyLineOperatorRange(row, endRow)
}

func (e *Engine) applyLineOperatorRange(startRow, endRow int) {
	startRow = clamp(startRow, 0, e.buf.lineCount()-1)
	endRow = clamp(endRow, 0, e.buf.lineCount()-1)
	if endRow < startRow {
		startRow, endRow = endRow, startRow
	}
	s := e.buf.line(startRow)
	for i := startRow + 1; i <= endRow; i++ {
		s += "\n" + e.buf.line(i)
	}
	e.register = s
	e.registerLinewise = true
	if e.pendingOperator == OperatorDelete || e.pendingOperator == OperatorChange {
		e.pushUndo()
		linesToDelete := endRow - startRow + 1
		for i := 0; i < linesToDelete && startRow < e.buf.lineCount(); i++ {
			e.buf.deleteLine(startRow)
		}
		e.cursor = e.normalizeNormalCursor(Cursor{Row: startRow, Col: 0})
		return
	}
	e.cursor = e.normalizeNormalCursor(Cursor{Row: startRow, Col: 0})
}

func (e *Engine) applyOperatorSelection(sel rangeSelection) {
	if e.pendingOperator == OperatorYank {
		e.register = e.slice(sel)
		e.registerLinewise = false
		return
	}
	if e.pendingOperator == OperatorDelete {
		e.pushUndo()
		e.register = e.buf.deleteRange(sel)
		e.registerLinewise = false
		e.cursor = e.normalizeNormalCursor(sel.start)
		return
	}
	if e.pendingOperator == OperatorChange {
		e.pushUndo()
		e.register = e.buf.deleteRange(sel)
		e.registerLinewise = false
		e.cursor = e.normalizeNormalCursor(sel.start)
	}
}

func (e *Engine) slice(sel rangeSelection) string {
	start := e.buf.clampCursor(sel.start)
	end := e.buf.clampCursor(sel.end)
	if start.Row > end.Row || (start.Row == end.Row && start.Col > end.Col) {
		start, end = end, start
	}
	if start.Row == end.Row {
		line := e.buf.line(start.Row)
		return line[start.Col:end.Col]
	}
	out := e.buf.line(start.Row)[start.Col:]
	for row := start.Row + 1; row < end.Row; row++ {
		out += "\n" + e.buf.line(row)
	}
	out += "\n" + e.buf.line(end.Row)[:end.Col]
	return out
}

func (e *Engine) clearOperatorState() {
	e.pendingOperator = OperatorNone
	e.awaitingTextObject = 0
	e.awaitingOperatorG = false
	e.pendingOperatorG = 0
	e.countPrefix = 0
	e.operatorCount = 0
	e.awaitingReplace = false
	e.replaceCount = 0
}

func (e *Engine) applyOperatorToLineTarget(targetRow int) {
	startRow := e.cursor.Row
	if targetRow < startRow {
		startRow, targetRow = targetRow, startRow
	}
	startRow = clamp(startRow, 0, e.buf.lineCount()-1)
	targetRow = clamp(targetRow, 0, e.buf.lineCount()-1)
	e.applyLineOperatorRange(startRow, targetRow)
}

func (e *Engine) deleteCharUnderCursor() {
	line := e.buf.line(e.cursor.Row)
	if len(line) == 0 || e.cursor.Col >= len(line) {
		return
	}
	e.pushUndo()
	sel := rangeSelection{start: e.cursor, end: Cursor{Row: e.cursor.Row, Col: e.cursor.Col + 1}}
	e.register = e.buf.deleteRange(sel)
	e.registerLinewise = false
	e.cursor = e.normalizeNormalCursor(e.cursor)
}

func (e *Engine) pasteAfter() {
	if e.register == "" && !e.registerLinewise {
		return
	}
	e.pushUndo()
	e.cursor = e.buf.insertAfter(e.cursor, e.register, e.registerLinewise)
	e.cursor = e.normalizeNormalCursor(e.cursor)
}

func (e *Engine) pasteBefore() {
	if e.register == "" && !e.registerLinewise {
		return
	}
	e.pushUndo()
	e.cursor = e.buf.insertBefore(e.cursor, e.register, e.registerLinewise)
	e.cursor = e.normalizeNormalCursor(e.cursor)
}

func (e *Engine) replaceUnderCursor(r rune, count int) {
	line := e.buf.line(e.cursor.Row)
	if len(line) == 0 || e.cursor.Col >= len(line) {
		return
	}
	if count <= 0 {
		count = 1
	}
	e.pushUndo()
	start := e.cursor
	end := Cursor{Row: start.Row, Col: start.Col + count}
	if end.Col > len(line) {
		end.Col = len(line)
	}
	replaceLen := end.Col - start.Col
	if replaceLen <= 0 {
		return
	}
	_ = e.buf.deleteRange(rangeSelection{start: start, end: end})
	c := start
	for i := 0; i < replaceLen; i++ {
		c = e.buf.insertRune(c, r)
	}
	e.cursor = e.normalizeNormalCursor(start)
}

func (e *Engine) undo() {
	if len(e.undoStack) == 0 {
		return
	}
	last := e.undoStack[len(e.undoStack)-1]
	e.undoStack = e.undoStack[:len(e.undoStack)-1]
	e.buf = newTextBuffer(last.lines)
	e.cursor = last.cursor
	e.mode = last.mode
	e.pendingOperator = last.pendingOperator
	e.register = last.register
	e.registerLinewise = last.registerLinewise
	e.awaitingTextObject = 0
	e.countPrefix = 0
	e.operatorCount = 0
	e.cursor = e.normalizeNormalCursor(e.cursor)
}

func (e *Engine) consumeCount() int {
	if e.countPrefix <= 0 {
		return 1
	}
	v := e.countPrefix
	e.countPrefix = 0
	return v
}

func (e *Engine) pushUndo() {
	e.undoStack = append(e.undoStack, snapshot{
		lines:            e.buf.cloneLines(),
		cursor:           e.cursor,
		mode:             e.mode,
		pendingOperator:  e.pendingOperator,
		register:         e.register,
		registerLinewise: e.registerLinewise,
	})
}

func (e *Engine) normalizeNormalCursor(c Cursor) Cursor {
	c = e.buf.clampCursor(c)
	lineLen := e.buf.lineLen(c.Row)
	if lineLen == 0 {
		c.Col = 0
		return c
	}
	if c.Col >= lineLen {
		c.Col = lineLen - 1
	}
	if c.Col < 0 {
		c.Col = 0
	}
	return c
}

func isWord(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func (e *Engine) globalIndex(c Cursor) int {
	idx := 0
	for row := 0; row < c.Row; row++ {
		idx += len(e.buf.line(row)) + 1
	}
	return idx + c.Col
}

func (e *Engine) cursorFromGlobal(idx int) Cursor {
	if idx <= 0 {
		return Cursor{}
	}
	remaining := idx
	for row := 0; row < e.buf.lineCount(); row++ {
		lineLen := len(e.buf.line(row))
		if remaining <= lineLen {
			return Cursor{Row: row, Col: remaining}
		}
		remaining -= lineLen
		if row < e.buf.lineCount()-1 {
			if remaining == 0 {
				return Cursor{Row: row + 1, Col: 0}
			}
			remaining--
		}
	}
	last := e.buf.lineCount() - 1
	if last < 0 {
		return Cursor{}
	}
	return Cursor{Row: last, Col: len(e.buf.line(last))}
}

func (e *Engine) contentRunes() []rune {
	text := ""
	for i := 0; i < e.buf.lineCount(); i++ {
		if i > 0 {
			text += "\n"
		}
		text += e.buf.line(i)
	}
	return []rune(text)
}

func (e *Engine) motionWordForward(c Cursor) Cursor {
	runes := e.contentRunes()
	idx := e.globalIndex(e.buf.clampCursor(c))
	if idx >= len(runes) {
		return c
	}
	i := idx
	if i < len(runes) && isWord(runes[i]) {
		for i < len(runes) && isWord(runes[i]) {
			i++
		}
	}
	for i < len(runes) && !isWord(runes[i]) {
		i++
	}
	if i >= len(runes) {
		return e.cursorFromGlobal(len(runes))
	}
	return e.cursorFromGlobal(i)
}

func (e *Engine) motionWordForwardExclusive(c Cursor) Cursor {
	next := e.motionWordForward(c)
	return e.buf.clampCursor(next)
}

func (e *Engine) motionWordBackward(c Cursor) Cursor {
	runes := e.contentRunes()
	idx := e.globalIndex(e.buf.clampCursor(c))
	if idx <= 0 {
		return Cursor{}
	}
	i := idx - 1
	for i >= 0 && !isWord(runes[i]) {
		i--
	}
	if i < 0 {
		return Cursor{}
	}
	for i >= 0 && isWord(runes[i]) {
		i--
	}
	return e.cursorFromGlobal(i + 1)
}

func (e *Engine) motionWordEnd(c Cursor) Cursor {
	runes := e.contentRunes()
	idx := e.globalIndex(e.buf.clampCursor(c))
	if idx >= len(runes) {
		if len(runes) == 0 {
			return Cursor{}
		}
		return e.cursorFromGlobal(len(runes) - 1)
	}
	i := idx
	if !isWord(runes[i]) {
		for i < len(runes) && !isWord(runes[i]) {
			i++
		}
		if i >= len(runes) {
			return e.cursorFromGlobal(len(runes) - 1)
		}
	}
	for i < len(runes) && isWord(runes[i]) {
		i++
	}
	return e.cursorFromGlobal(i - 1)
}

func (e *Engine) innerWordSelection(c Cursor, count int) rangeSelection {
	runes := e.contentRunes()
	if len(runes) == 0 {
		return rangeSelection{start: Cursor{}, end: Cursor{}}
	}
	idx := e.globalIndex(e.buf.clampCursor(c))
	if idx >= len(runes) {
		idx = len(runes) - 1
	}
	if !isWord(runes[idx]) {
		i := idx
		for i < len(runes) && !isWord(runes[i]) {
			i++
		}
		if i >= len(runes) {
			return rangeSelection{start: e.cursorFromGlobal(idx), end: e.cursorFromGlobal(idx)}
		}
		idx = i
	}
	start := idx
	for start > 0 && isWord(runes[start-1]) {
		start--
	}
	end := idx
	for end < len(runes) && isWord(runes[end]) {
		end++
	}
	for n := 1; n < count; n++ {
		i := end
		for i < len(runes) && !isWord(runes[i]) {
			i++
		}
		for i < len(runes) && isWord(runes[i]) {
			i++
		}
		end = i
	}
	return rangeSelection{start: e.cursorFromGlobal(start), end: e.cursorFromGlobal(end)}
}

func (e *Engine) aroundWordSelection(c Cursor, count int) rangeSelection {
	sel := e.innerWordSelection(c, count)
	runes := e.contentRunes()
	start := e.globalIndex(sel.start)
	end := e.globalIndex(sel.end)
	for end < len(runes) && (runes[end] == ' ' || runes[end] == '\t') {
		end++
	}
	if end == e.globalIndex(sel.end) {
		for start > 0 && (runes[start-1] == ' ' || runes[start-1] == '\t') {
			start--
		}
	}
	return rangeSelection{start: e.cursorFromGlobal(start), end: e.cursorFromGlobal(end)}
}
