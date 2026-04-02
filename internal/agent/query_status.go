package agent

import "github.com/alliecatowo/alliecode/internal/types"

// StatusView returns a typed runtime and replay status snapshot.
func (a *Agent) StatusView() types.AgentStatusView {
	cursor := a.ReplayCursor()
	a.eventMu.RLock()
	lifecycle := a.lifecycleState
	a.eventMu.RUnlock()

	return types.AgentStatusView{
		Lifecycle: lifecycle,
		Runtime:   a.RuntimeSnapshot(),
		Replay:    cursor,
		Window: types.AgentReplayWindow{
			Start: cursor.Start,
			End:   cursor.End,
			Total: cursor.Size,
		},
	}
}
