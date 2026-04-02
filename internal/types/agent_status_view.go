package types

// AgentStatusView combines runtime and replay metadata for status introspection.
type AgentStatusView struct {
	Lifecycle AgentLifecycleState  `json:"lifecycle,omitempty"`
	Runtime   AgentRuntimeSnapshot `json:"runtime,omitempty"`
	Replay    AgentReplayCursor    `json:"replay,omitempty"`
	Window    AgentReplayWindow    `json:"window,omitempty"`
}
