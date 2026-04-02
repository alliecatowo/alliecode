package providers

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

type introspectionProvider struct{}

func (introspectionProvider) Name() string { return "i" }
func (introspectionProvider) Chat(context.Context, types.ChatRequest) (<-chan types.StreamEvent, error) {
	ch := make(chan types.StreamEvent)
	close(ch)
	return ch, nil
}
func (introspectionProvider) ChatSync(context.Context, types.ChatRequest) (*types.ChatResponse, error) {
	return &types.ChatResponse{}, nil
}
func (introspectionProvider) ListModels(context.Context) ([]types.Model, error) {
	return []types.Model{{ID: "m1", Provider: "i", ContextWindow: 1000, MaxOutput: 100, SupportsTools: true}}, nil
}
func (introspectionProvider) SupportsStreaming() bool { return true }
func (introspectionProvider) SupportsTools() bool     { return true }
func (introspectionProvider) SupportsThinking() bool  { return false }

func TestGetProviderModelMetadataFallback(t *testing.T) {
	meta, err := GetProviderModelMetadata(context.Background(), introspectionProvider{})
	if err != nil {
		t.Fatalf("GetProviderModelMetadata error: %v", err)
	}
	if len(meta) != 1 || meta[0].Model != "m1" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
}
