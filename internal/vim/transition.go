package vim

// Transition captures a single state-machine key transition.
type Transition struct {
	Before State
	After  State
	Key    Key
}

// HandleKeyWithTransition applies one key and returns before/after states.
func (e *Engine) HandleKeyWithTransition(k Key) Transition {
	before := e.State()
	after := e.HandleKey(k)
	return Transition{Before: before, After: after, Key: k}
}

// HandleKeysWithTransitions applies parsed keys and records transitions.
func (e *Engine) HandleKeysWithTransitions(input string) []Transition {
	keys := ParseKeys(input)
	if len(keys) == 0 {
		return nil
	}
	out := make([]Transition, 0, len(keys))
	for _, key := range keys {
		out = append(out, e.HandleKeyWithTransition(key))
	}
	return out
}
