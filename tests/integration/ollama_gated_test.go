package integration_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/providers/ollama"
	"github.com/alliecatowo/alliecode/internal/types"
	"github.com/alliecatowo/alliecode/tests/integration/fixtures"
)

func TestOllamaListModels_Gated(t *testing.T) {
	fixtures.RequireEnv(t, "AC_IT_OLLAMA")

	p, err := ollama.New("", nil)
	if err != nil {
		t.Fatalf("ollama.New() error = %v", err)
	}

	models, err := p.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels() error = %v", err)
	}
	if len(models) == 0 {
		t.Fatalf("expected at least one local model from ollama")
	}
	for i, model := range models {
		if model.ID == "" {
			t.Fatalf("models[%d].ID is empty", i)
		}
		if model.Provider != "ollama" {
			t.Fatalf("models[%d].Provider = %q, want %q", i, model.Provider, "ollama")
		}
	}
}

func TestOllamaChatSyncSimplePrompt_Gated(t *testing.T) {
	fixtures.RequireEnv(t, "AC_IT_OLLAMA")

	p, err := ollama.New("", nil)
	if err != nil {
		t.Fatalf("ollama.New() error = %v", err)
	}

	temp := 0.0
	resp, err := p.ChatSync(context.Background(), types.ChatRequest{
		Model:       ollamaIntegrationModel(),
		Temperature: &temp,
		Messages: []types.Message{
			types.NewTextMessage(types.RoleUser, "Reply with exactly: integration-ok"),
		},
	})
	if err != nil {
		t.Fatalf("ChatSync() error = %v", err)
	}
	if got := strings.TrimSpace(resp.Message.GetText()); got == "" {
		t.Fatalf("response text is empty")
	}
}

func TestOllamaToolWorkflowFallback_Gated(t *testing.T) {
	fixtures.RequireEnv(t, "AC_IT_OLLAMA")

	p, err := ollama.New("", nil)
	if err != nil {
		t.Fatalf("ollama.New() error = %v", err)
	}

	temp := 0.0
	first, err := p.ChatSync(context.Background(), types.ChatRequest{
		Model:       ollamaIntegrationModel(),
		Temperature: &temp,
		Messages: []types.Message{
			types.NewTextMessage(types.RoleUser, "Use the add tool for 2 and 3, then answer with the total."),
		},
		Tools: []types.ToolDef{
			{
				Name:        "add",
				Description: "Add two integers",
				InputSchema: types.ToolSchema{
					Type: "object",
					Properties: map[string]types.PropertySchema{
						"a": {Type: "integer"},
						"b": {Type: "integer"},
					},
					Required: []string{"a", "b"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("first ChatSync() error = %v", err)
	}

	toolUses := first.Message.GetToolUses()
	if len(toolUses) == 0 {
		if got := strings.TrimSpace(first.Message.GetText()); got == "" {
			t.Fatalf("fallback response text is empty")
		}
		return
	}

	if toolUses[0].ToolUseID == "" {
		t.Fatalf("first tool use has empty tool_use_id")
	}

	toolResult := types.NewToolResultMessage(toolUses[0].ToolUseID, `{"total":5}`, false)
	second, err := p.ChatSync(context.Background(), types.ChatRequest{
		Model:       ollamaIntegrationModel(),
		Temperature: &temp,
		Messages: []types.Message{
			types.NewTextMessage(types.RoleUser, "Use the add tool for 2 and 3, then answer with the total."),
			first.Message,
			toolResult,
		},
		Tools: []types.ToolDef{
			{
				Name:        "add",
				Description: "Add two integers",
				InputSchema: types.ToolSchema{
					Type: "object",
					Properties: map[string]types.PropertySchema{
						"a": {Type: "integer"},
						"b": {Type: "integer"},
					},
					Required: []string{"a", "b"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("second ChatSync() error = %v", err)
	}

	if got := strings.TrimSpace(second.Message.GetText()); got == "" {
		payload, _ := json.Marshal(second.Message)
		t.Fatalf("second response text is empty: %s", string(payload))
	}
}

func ollamaIntegrationModel() string {
	if model := strings.TrimSpace(os.Getenv("AC_IT_OLLAMA_MODEL")); model != "" {
		return model
	}
	return "llama3"
}
