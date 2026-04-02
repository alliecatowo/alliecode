// Package ollama implements the Ollama (local models) provider for AllieCode.
// Ollama exposes an OpenAI-compatible API, so this wraps the openaicompat provider.
package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alliecatowo/alliecode/internal/providers/common"
	"github.com/alliecatowo/alliecode/internal/providers/openaicompat"
	"github.com/alliecatowo/alliecode/internal/types"
)

const defaultModel = "llama3"

// Provider implements types.Provider for local Ollama models.
type Provider struct {
	baseURL string
	inner   *openaicompat.Provider
	options map[string]string
}

// New creates a new Ollama provider.
func New(baseURL string, options map[string]string) (*Provider, error) {
	baseURL = common.ResolveAuthValue("OLLAMA_BASE_URL", baseURL, "http://localhost:11434")

	// Ollama exposes OpenAI-compatible endpoint at /v1
	inner, err := openaicompat.New("", baseURL+"/v1", options)
	if err != nil {
		return nil, err
	}

	return &Provider{
		baseURL: strings.TrimRight(baseURL, "/"),
		inner:   inner,
		options: options,
	}, nil
}

func (p *Provider) Name() string            { return "ollama" }
func (p *Provider) SupportsStreaming() bool { return true }
func (p *Provider) SupportsTools() bool     { return true }
func (p *Provider) SupportsThinking() bool  { return false }

func (p *Provider) RetryPolicy() types.ProviderRetryPolicy {
	return openaicompat.GetOpenAICompatRetryPolicy(p.inner)
}

func (p *Provider) ModelMetadata(ctx context.Context) ([]types.ModelMetadata, error) {
	models, err := p.ListModels(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]types.ModelMetadata, 0, len(models))
	for _, m := range models {
		out = append(out, types.ModelMetadata{
			Provider:        m.Provider,
			Model:           m.ID,
			ContextWindow:   m.ContextWindow,
			MaxOutput:       m.MaxOutput,
			SupportsText:    true,
			SupportsTools:   m.SupportsTools,
			SupportsToolUse: m.SupportsTools,
		})
	}
	return out, nil
}

// Chat streams a response from the local Ollama instance.
func (p *Provider) Chat(ctx context.Context, req types.ChatRequest) (<-chan types.StreamEvent, error) {
	if req.Model == "" {
		req.Model = defaultModel
	}
	return p.inner.Chat(ctx, req)
}

// ChatSync sends a synchronous request to Ollama.
func (p *Provider) ChatSync(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error) {
	if req.Model == "" {
		req.Model = defaultModel
	}
	return p.inner.ChatSync(ctx, req)
}

// ListModels queries the local Ollama instance for available models.
func (p *Provider) ListModels(ctx context.Context) ([]types.Model, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, common.TranslateError("ollama", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, common.TranslateError("ollama", fmt.Errorf("cannot connect to %s: %w", p.baseURL, err))
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, common.TranslateHTTPError("ollama", resp.StatusCode, body)
	}

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Models []struct {
			Name       string `json:"name"`
			ModifiedAt string `json:"modified_at"`
			Size       int64  `json:"size"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, common.TranslateError("ollama", err)
	}

	models := make([]types.Model, len(result.Models))
	for i, m := range result.Models {
		models[i] = types.Model{
			ID:            m.Name,
			Name:          m.Name,
			Provider:      "ollama",
			ContextWindow: 8192, // default, varies by model
			SupportsTools: true,
		}
	}
	return models, nil
}
