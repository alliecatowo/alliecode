// Package openaicompat implements a generic OpenAI-compatible provider.
// Works with vLLM, LMStudio, Together, Groq, and any other OpenAI-compatible API.
package openaicompat

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alliecatowo/alliecode/internal/providers/common"
	"github.com/alliecatowo/alliecode/internal/types"
)

// Provider implements types.Provider for any OpenAI-compatible API.
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
	options map[string]string
	retries common.DefaultRetryPolicy
}

// GetOpenAICompatRetryPolicy returns retry policy details for a compat provider.
func GetOpenAICompatRetryPolicy(p *Provider) types.ProviderRetryPolicy {
	if p == nil {
		return types.ProviderRetryPolicy{}
	}
	return p.RetryPolicy()
}

// New creates a new OpenAI-compatible provider.
func New(apiKey, baseURL string, options map[string]string) (*Provider, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("openai-compat: base_url is required")
	}
	return &Provider{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
		options: options,
		retries: common.NewDefaultRetryPolicy(),
	}, nil
}

func (p *Provider) Name() string            { return "openai-compat" }
func (p *Provider) SupportsStreaming() bool { return true }
func (p *Provider) SupportsTools() bool     { return true }
func (p *Provider) SupportsThinking() bool  { return false }

func (p *Provider) RetryPolicy() types.ProviderRetryPolicy {
	return types.ProviderRetryPolicy{
		Name:          "default",
		MaxRetries:    p.retries.MaxRetries,
		BaseBackoffMS: int(p.retries.BaseBackoff.Milliseconds()),
		MaxBackoffMS:  int(p.retries.MaxBackoff.Milliseconds()),
	}
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
			SupportsTools:   true,
			SupportsToolUse: true,
		})
	}
	return out, nil
}

// Chat sends a streaming request to the OpenAI-compatible endpoint.
func (p *Provider) Chat(ctx context.Context, req types.ChatRequest) (<-chan types.StreamEvent, error) {
	body := p.buildRequest(req, true)
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, common.TranslateError("openai-compat", err)
	}

	for attempt := 0; ; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
		if err != nil {
			return nil, common.TranslateError("openai-compat", err)
		}
		p.setHeaders(httpReq)

		resp, err := p.client.Do(httpReq)
		if err != nil {
			normalized := common.TranslateError("openai-compat", err)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("openai-compat", waitErr)
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			normalized := common.TranslateHTTPError("openai-compat", resp.StatusCode, respBody)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("openai-compat", waitErr)
			}
			continue
		}

		ch := make(chan types.StreamEvent, 64)
		go p.streamResponse(resp.Body, ch)
		return ch, nil
	}
}

// ChatSync sends a non-streaming request.
func (p *Provider) ChatSync(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error) {
	body := p.buildRequest(req, false)
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, common.TranslateError("openai-compat", err)
	}

	for attempt := 0; ; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
		if err != nil {
			return nil, common.TranslateError("openai-compat", err)
		}
		p.setHeaders(httpReq)

		resp, err := p.client.Do(httpReq)
		if err != nil {
			normalized := common.TranslateError("openai-compat", err)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("openai-compat", waitErr)
			}
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			normalized := common.TranslateHTTPError("openai-compat", resp.StatusCode, respBody)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("openai-compat", waitErr)
			}
			continue
		}

		var apiResp chatCompletionResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			return nil, common.TranslateError("openai-compat", err)
		}

		return convertResponse(&apiResp), nil
	}
}

// ListModels returns available models from the endpoint.
func (p *Provider) ListModels(ctx context.Context) ([]types.Model, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	if err != nil {
		return nil, err
	}
	p.setHeaders(req)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, common.TranslateError("openai-compat", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, common.TranslateHTTPError("openai-compat", resp.StatusCode, body)
	}

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, common.TranslateError("openai-compat", err)
	}

	models := make([]types.Model, len(result.Data))
	for i, m := range result.Data {
		models[i] = types.Model{
			ID:       m.ID,
			Name:     m.ID,
			Provider: "openai-compat",
		}
	}
	return models, nil
}

// --- Internal types (OpenAI chat completions format) ---

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	Tools       []chatTool    `json:"tools,omitempty"`
	Temperature *float64      `json:"temperature,omitempty"`
}

type chatMessage struct {
	Role       string     `json:"role"`
	Content    any        `json:"content,omitempty"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type chatTool struct {
	Type     string       `json:"type"`
	Function chatFunction `json:"function"`
}

type chatFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type toolCall struct {
	Index    int    `json:"index,omitempty"`
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type chatCompletionResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message      chatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content   string     `json:"content,omitempty"`
			ToolCalls []toolCall `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage,omitempty"`
}

// --- Conversion ---

func (p *Provider) buildRequest(req types.ChatRequest, stream bool) chatCompletionRequest {
	apiReq := chatCompletionRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Stream:      stream,
		Temperature: req.Temperature,
	}

	if req.System != "" {
		apiReq.Messages = append(apiReq.Messages, chatMessage{
			Role:    "system",
			Content: req.System,
		})
	}

	for _, msg := range req.Messages {
		toolUses := msg.GetToolUses()
		if len(toolUses) > 0 && msg.Role == types.RoleAssistant {
			apiMsg := chatMessage{Role: "assistant"}
			text := msg.GetText()
			if text != "" {
				apiMsg.Content = text
			}
			for _, tu := range toolUses {
				apiMsg.ToolCalls = append(apiMsg.ToolCalls, toolCall{
					ID:   tu.ToolUseID,
					Type: "function",
					Function: struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					}{
						Name:      tu.ToolName,
						Arguments: string(tu.Input),
					},
				})
			}
			apiReq.Messages = append(apiReq.Messages, apiMsg)
			continue
		}

		for _, block := range msg.Content {
			if block.Type == types.ContentToolResult {
				apiReq.Messages = append(apiReq.Messages, chatMessage{
					Role:       "tool",
					Content:    block.Content,
					ToolCallID: block.ForToolUseID,
				})
			}
		}

		text := msg.GetText()
		if text != "" {
			apiReq.Messages = append(apiReq.Messages, chatMessage{
				Role:    string(msg.Role),
				Content: text,
			})
		}
	}

	for _, tool := range req.Tools {
		schema, _ := json.Marshal(tool.InputSchema)
		apiReq.Tools = append(apiReq.Tools, chatTool{
			Type: "function",
			Function: chatFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  schema,
			},
		})
	}

	return apiReq
}

func (p *Provider) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.apiKey)
	}
}

func (p *Provider) streamResponse(body io.ReadCloser, ch chan<- types.StreamEvent) {
	defer close(ch)
	defer body.Close()

	ch <- types.StreamEvent{Type: types.StreamStart}
	toolIDByIndex := map[int]string{}
	toolNameByIndex := map[int]string{}
	stopReason := types.StopReason("")

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: stopReason}
			return
		}

		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta
		if delta.Content != "" {
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: delta.Content}
		}
		for _, tc := range delta.ToolCalls {
			toolID := tc.ID
			if toolID == "" {
				toolID = toolIDByIndex[tc.Index]
			} else {
				toolIDByIndex[tc.Index] = toolID
			}

			toolName := tc.Function.Name
			if toolName == "" {
				toolName = toolNameByIndex[tc.Index]
			} else {
				toolNameByIndex[tc.Index] = toolName
			}

			if tc.Function.Name != "" || tc.ID != "" {
				ch <- types.StreamEvent{
					Type:      types.StreamToolUseStart,
					ToolUseID: toolID,
					ToolName:  toolName,
				}
			}
			if tc.Function.Arguments != "" {
				ch <- types.StreamEvent{Type: types.StreamToolUseDelta, ToolUseID: toolID, ToolName: toolName, Delta: tc.Function.Arguments}
			}
		}

		if chunk.Choices[0].FinishReason != nil {
			stopReason = mapFinishReason(*chunk.Choices[0].FinishReason)
			switch stopReason {
			case types.StopToolUse:
				for idx, toolID := range toolIDByIndex {
					ch <- types.StreamEvent{Type: types.StreamToolUseDone, ToolUseID: toolID, ToolName: toolNameByIndex[idx]}
				}
			}
			if chunk.Usage != nil {
				ch <- types.StreamEvent{
					Usage: &types.Usage{
						InputTokens:  chunk.Usage.PromptTokens,
						OutputTokens: chunk.Usage.CompletionTokens,
					},
					UsageCumulative: true,
					StopReason:      stopReason,
				}
			}
		}
	}
}

func convertResponse(resp *chatCompletionResponse) *types.ChatResponse {
	if len(resp.Choices) == 0 {
		return &types.ChatResponse{Model: resp.Model}
	}

	choice := resp.Choices[0]
	msg := types.Message{Role: types.RoleAssistant}

	if content, ok := choice.Message.Content.(string); ok && content != "" {
		msg.Content = append(msg.Content, types.ContentBlock{
			Type: types.ContentText,
			Text: content,
		})
	}

	for _, tc := range choice.Message.ToolCalls {
		msg.Content = append(msg.Content, types.ContentBlock{
			Type:      types.ContentToolUse,
			ToolUseID: tc.ID,
			ToolName:  tc.Function.Name,
			Input:     json.RawMessage(tc.Function.Arguments),
		})
	}

	stopReason := types.StopEndTurn
	if choice.FinishReason != "" {
		stopReason = mapFinishReason(choice.FinishReason)
	}

	return &types.ChatResponse{
		Message: msg,
		Usage: types.Usage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
		StopReason: stopReason,
		Model:      resp.Model,
	}
}

func mapFinishReason(reason string) types.StopReason {
	switch reason {
	case "length", "max_tokens":
		return types.StopMaxTokens
	case "tool_calls":
		return types.StopToolUse
	case "stop":
		return types.StopEndTurn
	case "content_filter":
		return types.StopContentFilter
	case "context_length_exceeded":
		return types.StopContextLimit
	case "refusal":
		return types.StopRefusal
	default:
		return types.StopUnknown
	}
}
