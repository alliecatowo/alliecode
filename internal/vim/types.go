package vim

// Mode describes the current editing mode.
type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeOperatorPending
)

// Operator is a pending or executed editing operator.
type Operator int

const (
	OperatorNone Operator = iota
	OperatorDelete
	OperatorYank
	OperatorChange
)

// Cursor is a zero-based buffer position.
type Cursor struct {
	Row int
	Col int
}

// State is a deterministic snapshot of the engine.
type State struct {
	Mode            Mode
	Cursor          Cursor
	PendingOperator Operator
	Lines           []string
}

// SpecialKey is a non-printing key.
type SpecialKey int

const (
	SpecialNone SpecialKey = iota
	SpecialEsc
	SpecialEnter
	SpecialBackspace
	SpecialDelete
	SpecialTab
	SpecialShiftTab
	SpecialLeft
	SpecialRight
	SpecialUp
	SpecialDown
	SpecialPageUp
	SpecialPageDown
	SpecialHome
	SpecialEnd
)

// Key is one parsed keystroke.
type Key struct {
	Rune    rune
	Special SpecialKey
}
