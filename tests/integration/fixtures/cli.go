package fixtures

import (
	"context"

	"github.com/alliecatowo/alliecode/internal/commands"
)

// CLIInvocationFixture dispatches slash commands against a deterministic runtime state.
type CLIInvocationFixture struct {
	ctx      context.Context
	Registry *commands.Registry
	State    *commands.RuntimeState
}

func NewCLIInvocationFixture(state *commands.RuntimeState) CLIInvocationFixture {
	if state == nil {
		state = &commands.RuntimeState{}
	}
	return CLIInvocationFixture{
		ctx:      context.Background(),
		Registry: commands.DefaultRegistry(),
		State:    state,
	}
}

func (f CLIInvocationFixture) Invoke(input string) (commands.Result, error) {
	return f.Registry.Dispatch(f.ctx, commands.Context{State: f.State}, input)
}
