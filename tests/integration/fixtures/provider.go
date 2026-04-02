package fixtures

import (
	"context"
	"sort"

	"github.com/alliecatowo/alliecode/internal/types"
)

// DeterministicProvider is a fixture provider for integration tests.
type DeterministicProvider struct {
	replies      map[string]string
	defaultReply string
	models       []types.Model
}

func NewDeterministicProvider(replies map[string]string, defaultReply string) *DeterministicProvider {
	repliesCopy := make(map[string]string, len(replies))
	for k, v := range replies {
		repliesCopy[k] = v
	}

	return &DeterministicProvider{
		replies:      repliesCopy,
		defaultReply: defaultReply,
		models: []types.Model{
			{
				ID:            "fixture-deterministic-v1",
				Name:          "fixture-deterministic-v1",
				Provider:      "fixture",
				ContextWindow: 4096,
				MaxOutput:     512,
				SupportsTools: true,
			},
		},
	}
}

func (p *DeterministicProvider) Name() string            { return "fixture" }
func (p *DeterministicProvider) SupportsStreaming() bool { return true }
func (p *DeterministicProvider) SupportsTools() bool     { return true }
func (p *DeterministicProvider) SupportsThinking() bool  { return false }

func (p *DeterministicProvider) Chat(ctx context.Context, req types.ChatRequest) (<-chan types.StreamEvent, error) {
	out := make(chan types.StreamEvent, 3)
	reply := p.resolveReply(req)

	go func() {
		defer close(out)
		select {
		case <-ctx.Done():
			return
		case out <- types.StreamEvent{Type: types.StreamStart}:
		}

		select {
		case <-ctx.Done():
			return
		case out <- types.StreamEvent{Type: types.StreamContentDelta, Delta: reply}:
		}

		msg := types.NewTextMessage(types.RoleAssistant, reply)
		select {
		case <-ctx.Done():
			return
		case out <- types.StreamEvent{Type: types.StreamMessageDone, Message: &msg}:
		}
	}()

	return out, nil
}

func (p *DeterministicProvider) ChatSync(_ context.Context, req types.ChatRequest) (*types.ChatResponse, error) {
	reply := p.resolveReply(req)
	return &types.ChatResponse{
		Message:    types.NewTextMessage(types.RoleAssistant, reply),
		StopReason: types.StopEndTurn,
		Model:      "fixture-deterministic-v1",
		Usage: types.Usage{
			InputTokens:  1,
			OutputTokens: 1,
		},
	}, nil
}

func (p *DeterministicProvider) ListModels(_ context.Context) ([]types.Model, error) {
	models := make([]types.Model, len(p.models))
	copy(models, p.models)
	sort.Slice(models, func(i, j int) bool {
		return models[i].ID < models[j].ID
	})
	return models, nil
}

func (p *DeterministicProvider) resolveReply(req types.ChatRequest) string {
	lastUser := ""
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == types.RoleUser {
			lastUser = req.Messages[i].GetText()
			break
		}
	}

	if v, ok := p.replies[lastUser]; ok {
		return v
	}
	return p.defaultReply
}
