package agent

import (
	"context"
	"fmt"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

type summaryProvider struct{}

func (summaryProvider) Name() string { return "summary" }

func (summaryProvider) Chat(context.Context, types.ChatRequest) (<-chan types.StreamEvent, error) {
	ch := make(chan types.StreamEvent)
	close(ch)
	return ch, nil
}

func (summaryProvider) ChatSync(context.Context, types.ChatRequest) (*types.ChatResponse, error) {
	return &types.ChatResponse{Message: types.NewTextMessage(types.RoleAssistant, "summary")}, nil
}

func (summaryProvider) ListModels(context.Context) ([]types.Model, error) { return nil, nil }
func (summaryProvider) SupportsStreaming() bool                           { return true }
func (summaryProvider) SupportsTools() bool                               { return true }
func (summaryProvider) SupportsThinking() bool                            { return true }

func TestCompactMessagesBoundaryMatchesCompacted(t *testing.T) {
	msgs := []types.Message{
		types.NewTextMessage(types.RoleUser, "first"),
		types.NewTextMessage(types.RoleAssistant, "a1"),
		types.NewTextMessage(types.RoleUser, "u2"),
		types.NewTextMessage(types.RoleAssistant, "a2"),
		types.NewTextMessage(types.RoleUser, "u3"),
		types.NewTextMessage(types.RoleAssistant, "a3"),
		types.NewTextMessage(types.RoleUser, "u4"),
	}

	compacted, boundary, err := CompactMessages(context.Background(), msgs, summaryProvider{}, "test")
	if err != nil {
		t.Fatalf("CompactMessages() error = %v", err)
	}
	if boundary == nil {
		t.Fatalf("expected compaction boundary")
	}

	applied, err := applyBoundary(msgs, *boundary)
	if err != nil {
		t.Fatalf("applyBoundary() error = %v", err)
	}

	if len(applied) != len(compacted) {
		t.Fatalf("len(applied) = %d, len(compacted) = %d", len(applied), len(compacted))
	}
	for i := range compacted {
		if applied[i].Role != compacted[i].Role || applied[i].GetText() != compacted[i].GetText() {
			t.Fatalf("message %d mismatch after boundary apply", i)
		}
	}
}

func applyBoundary(messages []types.Message, boundary types.CompactionBoundary) ([]types.Message, error) {
	if boundary.StartIndex < 0 || boundary.EndIndex < boundary.StartIndex || boundary.EndIndex > len(messages) {
		return nil, fmt.Errorf("invalid boundary")
	}
	out := make([]types.Message, 0, len(messages)-(boundary.EndIndex-boundary.StartIndex)+len(boundary.Replacement))
	out = append(out, messages[:boundary.StartIndex]...)
	out = append(out, boundary.Replacement...)
	out = append(out, messages[boundary.EndIndex:]...)
	return out, nil
}
