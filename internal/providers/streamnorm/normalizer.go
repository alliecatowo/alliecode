package streamnorm

import "github.com/alliecatowo/alliecode/internal/types"

// Normalizer emits provider-agnostic stream events for loop/runtime.
type Normalizer struct {
	provider       string
	model          string
	started        bool
	usage          types.Usage
	activeToolID   string
	activeToolName string
	turnEnded      bool
}

// New creates a stream normalizer for one request.
func New(provider, model string) *Normalizer {
	return &Normalizer{provider: provider, model: model}
}

// Normalize returns normalized events, preserving source events while adding
// request/tool/usage boundary semantics.
func (n *Normalizer) Normalize(in types.StreamEvent) []types.StreamEvent {
	if n.turnEnded && in.Type != types.StreamError && in.Type != types.StreamUsageDelta && in.Type != types.StreamUsageTotal {
		return nil
	}

	out := make([]types.StreamEvent, 0, 5)
	if !n.started {
		n.started = true
		out = append(out, types.StreamEvent{
			Type:     types.StreamRequestStart,
			Provider: n.provider,
			Model:    n.model,
		})
	}

	if in.Type == types.StreamToolUseDelta && n.activeToolID == "" {
		synth := types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: in.ToolUseID, ToolName: in.ToolName}
		if synth.ToolUseID == "" {
			synth.ToolUseID = "tool-unknown"
		}
		out = append(out, n.normalizeSingle(synth)...)
	}
	if in.Type == types.StreamToolUseDone && n.activeToolID == "" {
		synth := types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: in.ToolUseID, ToolName: in.ToolName}
		if synth.ToolUseID == "" {
			synth.ToolUseID = "tool-unknown"
		}
		out = append(out, n.normalizeSingle(synth)...)
	}
	if in.Type == types.StreamContentDelta && n.activeToolID != "" {
		out = append(out, n.normalizeSingle(types.StreamEvent{Type: types.StreamToolUseDone, ToolUseID: n.activeToolID, ToolName: n.activeToolName})...)
	}

	out = append(out, n.normalizeSingle(in)...)
	return out
}

func (n *Normalizer) normalizeSingle(in types.StreamEvent) []types.StreamEvent {
	out := make([]types.StreamEvent, 0, 4)

	base := in
	base.Provider = n.provider
	if base.Model == "" {
		base.Model = n.model
	}
	originalUsage := base.Usage
	originalUsageCumulative := base.UsageCumulative
	base.Usage = nil
	base.UsageCumulative = false
	out = append(out, base)

	if in.Type == types.StreamToolUseStart {
		n.activeToolID = in.ToolUseID
		n.activeToolName = in.ToolName
		if n.activeToolID == "" {
			n.activeToolID = "tool-unknown"
		}
		out = append(out, types.StreamEvent{
			Type:      types.StreamToolBoundary,
			Boundary:  types.StreamBoundaryToolBegin,
			Provider:  n.provider,
			Model:     n.model,
			ToolUseID: n.activeToolID,
			ToolName:  in.ToolName,
		})
	}
	if in.Type == types.StreamToolUseDone {
		toolID := in.ToolUseID
		if toolID == "" {
			toolID = n.activeToolID
		}
		toolName := in.ToolName
		if toolName == "" {
			toolName = n.activeToolName
		}
		out = append(out, types.StreamEvent{
			Type:      types.StreamToolBoundary,
			Boundary:  types.StreamBoundaryToolEnd,
			Provider:  n.provider,
			Model:     n.model,
			ToolUseID: toolID,
			ToolName:  toolName,
		})
		n.activeToolID = ""
		n.activeToolName = ""
	}

	if originalUsage != nil {
		delta, total := n.advanceUsage(*originalUsage, originalUsageCumulative)
		if delta.InputTokens > 0 || delta.OutputTokens > 0 || delta.CacheHits > 0 {
			out = append(out, types.StreamEvent{
				Type:            types.StreamUsageDelta,
				Provider:        n.provider,
				Model:           n.model,
				Usage:           &delta,
				UsageCumulative: false,
				StopReason:      in.StopReason,
			})
		}
		out = append(out, types.StreamEvent{
			Type:            types.StreamUsageTotal,
			Provider:        n.provider,
			Model:           n.model,
			Usage:           &total,
			UsageCumulative: true,
			StopReason:      in.StopReason,
		})
	}

	if in.Type == types.StreamMessageDone {
		n.turnEnded = true
		if n.activeToolID != "" {
			out = append(out, types.StreamEvent{
				Type:      types.StreamToolBoundary,
				Boundary:  types.StreamBoundaryToolEnd,
				Provider:  n.provider,
				Model:     n.model,
				ToolUseID: n.activeToolID,
				ToolName:  n.activeToolName,
			})
			n.activeToolID = ""
			n.activeToolName = ""
		}
		out = append(out, types.StreamEvent{
			Type:       types.StreamToolBoundary,
			Boundary:   types.StreamBoundaryTurnEnd,
			Provider:   n.provider,
			Model:      n.model,
			StopReason: in.StopReason,
		})
	}

	return out
}

func (n *Normalizer) advanceUsage(usage types.Usage, cumulative bool) (types.Usage, types.Usage) {
	if !cumulative {
		n.usage.InputTokens += usage.InputTokens
		n.usage.OutputTokens += usage.OutputTokens
		n.usage.CacheHits += usage.CacheHits
		return usage, n.usage
	}

	delta := types.Usage{
		InputTokens:  usage.InputTokens - n.usage.InputTokens,
		OutputTokens: usage.OutputTokens - n.usage.OutputTokens,
		CacheHits:    usage.CacheHits - n.usage.CacheHits,
	}
	if usage.InputTokens < n.usage.InputTokens || usage.OutputTokens < n.usage.OutputTokens || usage.CacheHits < n.usage.CacheHits {
		delta = usage
	}
	if delta.InputTokens < 0 {
		delta.InputTokens = 0
	}
	if delta.OutputTokens < 0 {
		delta.OutputTokens = 0
	}
	if delta.CacheHits < 0 {
		delta.CacheHits = 0
	}
	n.usage = usage
	return delta, n.usage
}
