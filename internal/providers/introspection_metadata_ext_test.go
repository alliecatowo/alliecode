package providers

import (
	"context"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

type metadataEchoProvider struct{}

func (metadataEchoProvider) Name() string { return "echo" }
func (metadataEchoProvider) Chat(context.Context, types.ChatRequest) (<-chan types.StreamEvent, error) {
	return nil, nil
}
func (metadataEchoProvider) ChatSync(context.Context, types.ChatRequest) (*types.ChatResponse, error) {
	return nil, nil
}
func (metadataEchoProvider) ListModels(context.Context) ([]types.Model, error) {
	return []types.Model{{ID: "o3", Provider: "openai", ContextWindow: 200000, MaxOutput: 100000, SupportsTools: true, SupportsVision: true}}, nil
}
func (metadataEchoProvider) SupportsStreaming() bool { return true }
func (metadataEchoProvider) SupportsTools() bool     { return true }
func (metadataEchoProvider) SupportsThinking() bool  { return false }

func TestGetProviderModelMetadataIncludesTierAndCostClass(t *testing.T) {
	meta, err := GetProviderModelMetadata(context.Background(), metadataEchoProvider{})
	if err != nil {
		t.Fatalf("GetProviderModelMetadata() err = %v", err)
	}
	if len(meta) != 1 {
		t.Fatalf("expected single metadata entry")
	}
	if meta[0].Tier == "" || meta[0].CostClass == "" {
		t.Fatalf("expected tier and cost class from registry lookup, got %+v", meta[0])
	}
}
