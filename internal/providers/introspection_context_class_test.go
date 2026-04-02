package providers

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

type introspectionContextProvider struct{}

func (introspectionContextProvider) Name() string { return "fake" }

func (introspectionContextProvider) Chat(context.Context, types.ChatRequest) (<-chan types.StreamEvent, error) {
	return nil, nil
}

func (introspectionContextProvider) ChatSync(context.Context, types.ChatRequest) (*types.ChatResponse, error) {
	return nil, nil
}

func (introspectionContextProvider) ListModels(context.Context) ([]types.Model, error) {
	return []types.Model{{ID: "gpt-4o", Provider: "openai", ContextWindow: 128000, MaxOutput: 16384, SupportsTools: true}}, nil
}

func (introspectionContextProvider) SupportsStreaming() bool { return true }
func (introspectionContextProvider) SupportsTools() bool     { return true }
func (introspectionContextProvider) SupportsThinking() bool  { return false }

func TestGetProviderModelMetadataIncludesContextClass(t *testing.T) {
	meta, err := GetProviderModelMetadata(context.Background(), introspectionContextProvider{})
	if err != nil {
		t.Fatalf("GetProviderModelMetadata error: %v", err)
	}
	if len(meta) != 1 || meta[0].ContextClass == "" {
		t.Fatalf("expected context class in metadata: %+v", meta)
	}
}
