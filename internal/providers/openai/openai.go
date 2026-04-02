// Package openai implements the OpenAI provider for AllieCode.
package openai

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

const (
	defaultBaseURL = "https://api.openai.com/v1"
	defaultModel   = "gpt-4o"
)

// Provider implements types.Provider for the OpenAI API.
type Provider struct {
	apiKey  string
	baseURL string
	orgID   string
	client  *http.Client
	options map[string]string
	retries common.DefaultRetryPolicy
}

// New creates a new OpenAI provider.
func New(apiKey, baseURL, orgID string, options map[string]string) (*Provider, error) {
	apiKey = common.ResolveAuthValue("OPENAI_API_KEY", apiKey, "")
	if apiKey == "" {
		return nil, fmt.Errorf("openai: API key required (set OPENAI_API_KEY)")
	}
	baseURL = common.ResolveAuthValue("OPENAI_BASE_URL", baseURL, defaultBaseURL)

	return &Provider{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		orgID:   orgID,
		client:  &http.Client{},
		options: options,
		retries: common.NewDefaultRetryPolicy(),
	}, nil
}

func (p *Provider) Name() string            { return "openai" }
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
			Provider:            m.Provider,
			Model:               m.ID,
			ContextWindow:       m.ContextWindow,
			MaxOutput:           m.MaxOutput,
			SupportsText:        true,
			SupportsImage:       true,
			SupportsAudio:       true,
			SupportsTools:       m.SupportsTools,
			SupportsToolUse:     m.SupportsTools,
			SupportsThinking:    false,
			SupportsVision:      m.SupportsVision,
			SupportsAttachments: true,
		})
	}
	return out, nil
}

// Chat sends a streaming request to the OpenAI chat completions API.
func (p *Provider) Chat(ctx context.Context, req types.ChatRequest) (<-chan types.StreamEvent, error) {
	body := p.buildRequest(req, true)
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, common.TranslateError("openai", err)
	}

	for attempt := 0; ; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
		if err != nil {
			return nil, common.TranslateError("openai", err)
		}
		p.setHeaders(httpReq)

		resp, err := p.client.Do(httpReq)
		if err != nil {
			normalized := common.TranslateError("openai", err)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("openai", waitErr)
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			normalized := common.TranslateHTTPError("openai", resp.StatusCode, respBody)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("openai", waitErr)
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
		return nil, common.TranslateError("openai", err)
	}

	for attempt := 0; ; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(jsonBody))
		if err != nil {
			return nil, common.TranslateError("openai", err)
		}
		p.setHeaders(httpReq)

		resp, err := p.client.Do(httpReq)
		if err != nil {
			normalized := common.TranslateError("openai", err)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("openai", waitErr)
			}
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			normalized := common.TranslateHTTPError("openai", resp.StatusCode, respBody)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("openai", waitErr)
			}
			continue
		}

		var apiResp openAIResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			return nil, common.TranslateError("openai", err)
		}

		return p.convertResponse(&apiResp), nil
	}
}

// ListModels returns available OpenAI models.
func (p *Provider) ListModels(ctx context.Context) ([]types.Model, error) {
	return []types.Model{
		{ID: "gpt-4o", Name: "GPT-4o", Provider: "openai", ContextWindow: 128000, MaxOutput: 16384, SupportsTools: true, SupportsVision: true},
		{ID: "gpt-4o-mini", Name: "GPT-4o Mini", Provider: "openai", ContextWindow: 128000, MaxOutput: 16384, SupportsTools: true, SupportsVision: true},
		{ID: "o3", Name: "o3", Provider: "openai", ContextWindow: 200000, MaxOutput: 100000, SupportsTools: true, SupportsVision: true},
		{ID: "o4-mini", Name: "o4 Mini", Provider: "openai", ContextWindow: 200000, MaxOutput: 100000, SupportsTools: true, SupportsVision: true},
	}, nil
}

// --- Internal types ---

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Tools       []openAITool    `json:"tools,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    any              `json:"content,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	Name       string           `json:"name,omitempty"`
}

type openAITool struct {
	Type     string         `json:"type"`
	Function openAIFunction `json:"function"`
}

type openAIFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type openAIToolCall struct {
	Index    int    `json:"index,omitempty"`
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type openAIResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message      openAIMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

type openAIStreamChunk struct {
	ID      string `json:"id"`
	Choices []struct {
		Delta struct {
			Role      string           `json:"role,omitempty"`
			Content   string           `json:"content,omitempty"`
			ToolCalls []openAIToolCall `json:"tool_calls,omitempty"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage,omitempty"`
}

// --- Conversion ---

func (p *Provider) buildRequest(req types.ChatRequest, stream bool) openAIRequest {
	model := req.Model
	if model == "" {
		model = defaultModel
	}

	apiReq := openAIRequest{
		Model:       model,
		MaxTokens:   req.MaxTokens,
		Stream:      stream,
		Temperature: req.Temperature,
	}

	// System message
	if req.System != "" {
		apiReq.Messages = append(apiReq.Messages, openAIMessage{
			Role:    "system",
			Content: req.System,
		})
	}

	// Convert messages
	for _, msg := range req.Messages {
		toolUses := msg.GetToolUses()
		if len(toolUses) > 0 && msg.Role == types.RoleAssistant {
			// Assistant message with tool calls
			apiMsg := openAIMessage{Role: "assistant"}
			text := msg.GetText()
			if text != "" {
				apiMsg.Content = text
			}
			for _, tu := range toolUses {
				apiMsg.ToolCalls = append(apiMsg.ToolCalls, openAIToolCall{
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

		// Check for tool results
		for _, block := range msg.Content {
			if block.Type == types.ContentToolResult {
				apiReq.Messages = append(apiReq.Messages, openAIMessage{
					Role:       "tool",
					Content:    block.Content,
					ToolCallID: block.ForToolUseID,
				})
				continue
			}
		}

		// Regular text message
		text := msg.GetText()
		if text != "" {
			apiReq.Messages = append(apiReq.Messages, openAIMessage{
				Role:    string(msg.Role),
				Content: text,
			})
		}
	}

	// Convert tools
	for _, tool := range req.Tools {
		schema, _ := json.Marshal(tool.InputSchema)
		apiReq.Tools = append(apiReq.Tools, openAITool{
			Type: "function",
			Function: openAIFunction{
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
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	if p.orgID != "" {
		req.Header.Set("OpenAI-Organization", p.orgID)
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

		var chunk openAIStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]

		// Text content
		if choice.Delta.Content != "" {
			ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: choice.Delta.Content}
		}

		// Tool calls
		for _, tc := range choice.Delta.ToolCalls {
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

		if choice.FinishReason != nil {
			stopReason = mapFinishReason(*choice.FinishReason)
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

func (p *Provider) convertResponse(resp *openAIResponse) *types.ChatResponse {
	if len(resp.Choices) == 0 {
		return &types.ChatResponse{Model: resp.Model}
	}

	choice := resp.Choices[0]
	msg := types.Message{Role: types.RoleAssistant}

	// Text content
	if content, ok := choice.Message.Content.(string); ok && content != "" {
		msg.Content = append(msg.Content, types.ContentBlock{
			Type: types.ContentText,
			Text: content,
		})
	}

	// Tool calls
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
