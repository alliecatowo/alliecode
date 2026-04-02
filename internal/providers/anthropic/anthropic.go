// Package anthropic implements the Anthropic (Claude) provider for AllieCode.
package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/alliecatowo/alliecode/internal/providers/common"
	"github.com/alliecatowo/alliecode/internal/types"
)

const (
	defaultBaseURL   = "https://api.anthropic.com"
	defaultModel     = "claude-sonnet-4-20250514"
	apiVersion       = "2023-06-01"
	messagesEndpoint = "/v1/messages"
)

// Provider implements types.Provider for the Anthropic API.
type Provider struct {
	apiKey    string
	authToken string
	baseURL   string
	client    *http.Client
	options   map[string]string
	retries   common.DefaultRetryPolicy
}

// New creates a new Anthropic provider.
func New(apiKey, authToken, accessToken, baseURL string, options map[string]string) (*Provider, error) {
	apiKey = common.ResolveAuthValue("ANTHROPIC_API_KEY", apiKey, "")
	authToken = resolveAuthToken(authToken, accessToken)
	if apiKey == "" && authToken == "" {
		return nil, fmt.Errorf("anthropic: credentials required (set ANTHROPIC_AUTH_TOKEN/ANTHROPIC_ACCESS_TOKEN or ANTHROPIC_API_KEY)")
	}
	baseURL = common.ResolveAuthValue("ANTHROPIC_BASE_URL", baseURL, defaultBaseURL)

	return &Provider{
		apiKey:    apiKey,
		authToken: authToken,
		baseURL:   strings.TrimRight(baseURL, "/"),
		client:    &http.Client{},
		options:   options,
		retries:   common.NewDefaultRetryPolicy(),
	}, nil
}

func (p *Provider) Name() string            { return "anthropic" }
func (p *Provider) SupportsStreaming() bool { return true }
func (p *Provider) SupportsTools() bool     { return true }
func (p *Provider) SupportsThinking() bool  { return true }

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
			SupportsTools:       m.SupportsTools,
			SupportsToolUse:     m.SupportsTools,
			SupportsThinking:    true,
			SupportsVision:      m.SupportsVision,
			SupportsAttachments: true,
		})
	}
	return out, nil
}

// Chat sends a streaming chat request to the Anthropic API.
func (p *Provider) Chat(ctx context.Context, req types.ChatRequest) (<-chan types.StreamEvent, error) {
	body := p.buildRequest(req, true)

	for attempt := 0; ; attempt++ {
		httpReq, err := p.newRequest(ctx, body)
		if err != nil {
			return nil, common.TranslateError("anthropic", err)
		}

		resp, err := p.client.Do(httpReq)
		if err != nil {
			normalized := common.TranslateError("anthropic", err)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("anthropic", waitErr)
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			normalized := p.classifyAnthropicHTTPError(resp.StatusCode, resp.Header, respBody)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("anthropic", waitErr)
			}
			continue
		}

		ch := make(chan types.StreamEvent, 64)
		go p.streamResponse(resp.Body, ch)
		return ch, nil
	}
}

// ChatSync sends a non-streaming request and returns the complete response.
func (p *Provider) ChatSync(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error) {
	body := p.buildRequest(req, false)

	for attempt := 0; ; attempt++ {
		httpReq, err := p.newRequest(ctx, body)
		if err != nil {
			return nil, common.TranslateError("anthropic", err)
		}

		resp, err := p.client.Do(httpReq)
		if err != nil {
			normalized := p.classifyTransportError(err)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("anthropic", waitErr)
			}
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, common.TranslateError("anthropic", readErr)
		}

		if resp.StatusCode != http.StatusOK {
			normalized := p.classifyAnthropicHTTPError(resp.StatusCode, resp.Header, respBody)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("anthropic", waitErr)
			}
			continue
		}

		var apiResp anthropicResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			return nil, common.TranslateError("anthropic", err)
		}

		return p.convertResponse(&apiResp), nil
	}
}

// ListModels returns available Claude models.
func (p *Provider) ListModels(ctx context.Context) ([]types.Model, error) {
	return []types.Model{
		{ID: "claude-opus-4-20250514", Name: "Claude Opus 4", Provider: "anthropic", ContextWindow: 200000, MaxOutput: 32000, SupportsTools: true, SupportsVision: true},
		{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4", Provider: "anthropic", ContextWindow: 200000, MaxOutput: 64000, SupportsTools: true, SupportsVision: true},
		{ID: "claude-haiku-3-5-20241022", Name: "Claude Haiku 3.5", Provider: "anthropic", ContextWindow: 200000, MaxOutput: 8192, SupportsTools: true, SupportsVision: true},
	}, nil
}

// --- Internal types for Anthropic API ---

type anthropicRequest struct {
	Model     string             `json:"model"`
	Messages  []anthropicMessage `json:"messages"`
	System    string             `json:"system,omitempty"`
	MaxTokens int                `json:"max_tokens"`
	Stream    bool               `json:"stream,omitempty"`
	Tools     []anthropicTool    `json:"tools,omitempty"`
	Metadata  map[string]string  `json:"metadata,omitempty"`
}

type anthropicMessage struct {
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
}

type anthropicContent struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   string          `json:"content,omitempty"`
	IsError   bool            `json:"is_error,omitempty"`
}

type anthropicTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type anthropicResponse struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Role       string             `json:"role"`
	Content    []anthropicContent `json:"content"`
	Model      string             `json:"model"`
	StopReason string             `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

type anthropicStreamEvent struct {
	Type         string             `json:"type"`
	Index        int                `json:"index,omitempty"`
	Delta        json.RawMessage    `json:"delta,omitempty"`
	ContentBlock *anthropicContent  `json:"content_block,omitempty"`
	Message      *anthropicResponse `json:"message,omitempty"`
	Usage        *struct {
		OutputTokens int `json:"output_tokens"`
	} `json:"usage,omitempty"`
}

// --- Request/Response conversion ---

func (p *Provider) buildRequest(req types.ChatRequest, stream bool) anthropicRequest {
	model := req.Model
	if model == "" {
		model = defaultModel
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 16384
	}

	apiReq := anthropicRequest{
		Model:     model,
		System:    req.System,
		MaxTokens: maxTokens,
		Stream:    stream,
	}

	// Convert messages
	for _, msg := range req.Messages {
		apiMsg := anthropicMessage{Role: string(msg.Role)}
		for _, block := range msg.Content {
			switch block.Type {
			case types.ContentText:
				apiMsg.Content = append(apiMsg.Content, anthropicContent{
					Type: "text",
					Text: block.Text,
				})
			case types.ContentToolUse:
				apiMsg.Content = append(apiMsg.Content, anthropicContent{
					Type:  "tool_use",
					ID:    block.ToolUseID,
					Name:  block.ToolName,
					Input: block.Input,
				})
			case types.ContentToolResult:
				apiMsg.Content = append(apiMsg.Content, anthropicContent{
					Type:      "tool_result",
					ToolUseID: block.ForToolUseID,
					Content:   block.Content,
					IsError:   block.IsError,
				})
			case types.ContentThinking:
				apiMsg.Content = append(apiMsg.Content, anthropicContent{
					Type: "thinking",
					Text: block.Text,
				})
			}
		}
		apiReq.Messages = append(apiReq.Messages, apiMsg)
	}

	// Convert tools
	for _, tool := range req.Tools {
		schema, _ := json.Marshal(tool.InputSchema)
		apiReq.Tools = append(apiReq.Tools, anthropicTool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: schema,
		})
	}

	return apiReq
}

func (p *Provider) newRequest(ctx context.Context, body anthropicRequest) (*http.Request, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("anthropic: marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+messagesEndpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	header, value := p.authHeader()
	req.Header.Set(header, value)
	req.Header.Set("Anthropic-Version", apiVersion)
	req.Header.Set("Anthropic-Beta", "max-tokens-3-5-sonnet-2024-07-15")

	return req, nil
}

func (p *Provider) authHeader() (string, string) {
	if p.authToken != "" {
		return "Authorization", "Bearer " + p.authToken
	}
	return "X-API-Key", p.apiKey
}

func resolveAuthToken(configuredAuthToken, configuredAccessToken string) string {
	if v := os.Getenv("ANTHROPIC_AUTH_TOKEN"); v != "" {
		return v
	}
	if v := os.Getenv("ANTHROPIC_ACCESS_TOKEN"); v != "" {
		return v
	}
	if configuredAuthToken != "" {
		return configuredAuthToken
	}
	return configuredAccessToken
}

type anthropicErrorEnvelope struct {
	Type  string `json:"type"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
	Message string `json:"message"`
}

func (p *Provider) classifyTransportError(err error) error {
	normalized := common.TranslateError("anthropic", err)
	nerr, ok := normalized.(*common.NormalizedError)
	if !ok {
		return normalized
	}

	if nerr.Class != common.ErrorClassTransport {
		return normalized
	}

	nerr.Message = "transport error contacting Anthropic API; check network/proxy connectivity and TLS trust configuration"
	return nerr
}

func (p *Provider) classifyAnthropicHTTPError(statusCode int, headers http.Header, body []byte) error {
	base := common.TranslateHTTPError("anthropic", statusCode, body)
	nerr, ok := base.(*common.NormalizedError)
	if !ok {
		return base
	}

	errType, detail := parseAnthropicErrorPayload(body)
	detail = sanitizeDiagnostic(detail)
	errTypeLower := strings.ToLower(errType)
	detailLower := strings.ToLower(detail)
	requestID := strings.TrimSpace(headers.Get("request-id"))
	retryAfter := strings.TrimSpace(headers.Get("retry-after"))

	if isAnthropicModelNotFound(statusCode, errTypeLower, detailLower) {
		nerr.Class = common.ErrorClassModelNotFound
		nerr.Retryable = false
		nerr.Message = "model not found or not enabled for this account; verify the configured model name and entitlement"
		return attachHTTPDiagnostics(nerr, requestID, detail, "")
	}

	if isAnthropicAuthError(statusCode, errTypeLower, detailLower) {
		nerr.Class = common.ErrorClassAuth
		nerr.Retryable = false
		nerr.Message = "authentication failed; verify ANTHROPIC_API_KEY or ANTHROPIC_AUTH_TOKEN and refresh login if using OAuth"
		return attachHTTPDiagnostics(nerr, requestID, detail, "")
	}

	if isAnthropicQuotaError(statusCode, errTypeLower, detailLower) {
		nerr.Class = common.ErrorClassQuota
		nerr.Retryable = false
		nerr.Message = "quota exceeded; check Anthropic billing/usage limits and available credits"
		return attachHTTPDiagnostics(nerr, requestID, detail, "")
	}

	if statusCode == http.StatusTooManyRequests {
		nerr.Class = common.ErrorClassRateLimit
		nerr.Retryable = true
		nerr.Message = "rate limit reached; retry with backoff"
		return attachHTTPDiagnostics(nerr, requestID, detail, retryAfter)
	}

	if statusCode >= 500 {
		nerr.Message = "Anthropic service unavailable; retry shortly"
		return attachHTTPDiagnostics(nerr, requestID, detail, "")
	}

	if detail != "" {
		nerr.Message = detail
	}
	return attachHTTPDiagnostics(nerr, requestID, "", "")
}

func parseAnthropicErrorPayload(body []byte) (string, string) {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return "", ""
	}
	var payload anthropicErrorEnvelope
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", trimmed
	}
	errType := strings.TrimSpace(payload.Type)
	if payload.Error != nil {
		if strings.TrimSpace(payload.Error.Type) != "" {
			errType = strings.TrimSpace(payload.Error.Type)
		}
		if strings.TrimSpace(payload.Error.Message) != "" {
			return errType, payload.Error.Message
		}
	}
	if strings.TrimSpace(payload.Message) != "" {
		return errType, payload.Message
	}
	return errType, trimmed
}

func sanitizeDiagnostic(message string) string {
	clean := strings.TrimSpace(message)
	if clean == "" {
		return ""
	}
	clean = strings.Join(strings.Fields(clean), " ")
	for _, marker := range []string{"sk-ant-", "sk-"} {
		if idx := strings.Index(clean, marker); idx >= 0 {
			end := idx + len(marker)
			for end < len(clean) {
				ch := clean[end]
				if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' {
					end++
					continue
				}
				break
			}
			clean = clean[:idx] + "[redacted]" + clean[end:]
		}
	}
	return clean
}

func isAnthropicAuthError(statusCode int, errTypeLower, detailLower string) bool {
	if statusCode == http.StatusUnauthorized {
		return true
	}
	if statusCode == http.StatusForbidden {
		return strings.Contains(errTypeLower, "auth") || strings.Contains(detailLower, "oauth") || strings.Contains(detailLower, "token") || strings.Contains(detailLower, "x-api-key")
	}
	if strings.Contains(errTypeLower, "auth") {
		return true
	}
	return strings.Contains(detailLower, "invalid x-api-key") || strings.Contains(detailLower, "oauth token has been revoked")
}

func isAnthropicQuotaError(statusCode int, errTypeLower, detailLower string) bool {
	if statusCode != http.StatusTooManyRequests {
		return false
	}
	if strings.Contains(errTypeLower, "quota") || strings.Contains(errTypeLower, "credit") {
		return true
	}
	return strings.Contains(detailLower, "quota") || strings.Contains(detailLower, "credit balance is too low") || strings.Contains(detailLower, "insufficient credits")
}

func isAnthropicModelNotFound(statusCode int, errTypeLower, detailLower string) bool {
	if statusCode == http.StatusNotFound {
		return true
	}
	if statusCode != http.StatusBadRequest {
		return false
	}
	if strings.Contains(errTypeLower, "not_found") && strings.Contains(detailLower, "model") {
		return true
	}
	return strings.Contains(detailLower, "invalid model") || strings.Contains(detailLower, "model not found")
}

func attachHTTPDiagnostics(nerr *common.NormalizedError, requestID, detail, retryAfter string) *common.NormalizedError {
	parts := []string{nerr.Message}
	if detail != "" {
		parts = append(parts, fmt.Sprintf("detail=%s", detail))
	}
	if retryAfter != "" {
		parts = append(parts, fmt.Sprintf("retry_after=%ss", strings.TrimSpace(retryAfter)))
	}
	if requestID != "" {
		parts = append(parts, fmt.Sprintf("request_id=%s", requestID))
	}
	nerr.Message = strings.Join(parts, " | ")
	return nerr
}

func (p *Provider) streamResponse(body io.ReadCloser, ch chan<- types.StreamEvent) {
	defer close(ch)
	defer body.Close()

	blockTypes := map[int]string{}
	toolIDByIndex := map[int]string{}
	toolNameByIndex := map[int]string{}
	stopReason := types.StopReason("")

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			return
		}

		var event anthropicStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			ch <- types.StreamEvent{Type: types.StreamError, Error: err}
			return
		}

		switch event.Type {
		case "message_start":
			ch <- types.StreamEvent{Type: types.StreamStart}

		case "content_block_start":
			if event.ContentBlock != nil {
				blockTypes[event.Index] = event.ContentBlock.Type
				switch event.ContentBlock.Type {
				case "tool_use":
					toolIDByIndex[event.Index] = event.ContentBlock.ID
					toolNameByIndex[event.Index] = event.ContentBlock.Name
					ch <- types.StreamEvent{
						Type:      types.StreamToolUseStart,
						ToolUseID: event.ContentBlock.ID,
						ToolName:  event.ContentBlock.Name,
					}
				}
			}

		case "content_block_delta":
			if event.Delta != nil {
				var delta struct {
					Type        string `json:"type"`
					Text        string `json:"text"`
					PartialJSON string `json:"partial_json"`
					Thinking    string `json:"thinking"`
				}
				json.Unmarshal(event.Delta, &delta)

				switch delta.Type {
				case "text_delta":
					ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: delta.Text}
				case "input_json_delta":
					ch <- types.StreamEvent{Type: types.StreamToolUseDelta, ToolUseID: toolIDByIndex[event.Index], ToolName: toolNameByIndex[event.Index], Delta: delta.PartialJSON}
				case "thinking_delta":
					ch <- types.StreamEvent{Type: types.StreamThinkingDelta, Delta: delta.Thinking}
				}
			}

		case "content_block_stop":
			switch blockTypes[event.Index] {
			case "tool_use":
				ch <- types.StreamEvent{Type: types.StreamToolUseDone, ToolUseID: toolIDByIndex[event.Index], ToolName: toolNameByIndex[event.Index]}
			default:
				ch <- types.StreamEvent{Type: types.StreamContentDone}
			}
			delete(blockTypes, event.Index)
			delete(toolIDByIndex, event.Index)
			delete(toolNameByIndex, event.Index)

		case "message_delta":
			var delta struct {
				StopReason string `json:"stop_reason"`
			}
			_ = json.Unmarshal(event.Delta, &delta)
			if delta.StopReason != "" {
				stopReason = mapStopReason(delta.StopReason)
			}
			if event.Usage != nil {
				ch <- types.StreamEvent{
					Usage:           &types.Usage{OutputTokens: event.Usage.OutputTokens},
					UsageCumulative: true,
					StopReason:      stopReason,
				}
			}

		case "message_stop":
			ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: stopReason}
		}
	}
}

func (p *Provider) convertResponse(resp *anthropicResponse) *types.ChatResponse {
	msg := types.Message{Role: types.RoleAssistant}

	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			msg.Content = append(msg.Content, types.ContentBlock{
				Type: types.ContentText,
				Text: block.Text,
			})
		case "tool_use":
			msg.Content = append(msg.Content, types.ContentBlock{
				Type:      types.ContentToolUse,
				ToolUseID: block.ID,
				ToolName:  block.Name,
				Input:     block.Input,
			})
		case "thinking":
			msg.Content = append(msg.Content, types.ContentBlock{
				Type: types.ContentThinking,
				Text: block.Text,
			})
		}
	}

	stopReason := types.StopEndTurn
	if resp.StopReason != "" {
		stopReason = mapStopReason(resp.StopReason)
	}

	return &types.ChatResponse{
		Message: msg,
		Usage: types.Usage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
		},
		StopReason: stopReason,
		Model:      resp.Model,
	}
}

func mapStopReason(reason string) types.StopReason {
	switch reason {
	case "max_tokens":
		return types.StopMaxTokens
	case "model_context_window_exceeded":
		return types.StopContextLimit
	case "tool_use":
		return types.StopToolUse
	case "stop_sequence":
		return types.StopStopSequence
	case "end_turn":
		return types.StopEndTurn
	case "refusal":
		return types.StopRefusal
	default:
		return types.StopUnknown
	}
}
