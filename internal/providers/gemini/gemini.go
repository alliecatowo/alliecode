// Package gemini implements the Google Gemini provider for AllieCode.
package gemini

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/alliecatowo/alliecode/internal/providers/common"
	"github.com/alliecatowo/alliecode/internal/types"
)

const (
	defaultBaseURL = "https://generativelanguage.googleapis.com"
	defaultModel   = "gemini-2.5-pro"
	apiVersion     = "v1beta"
)

// Provider implements types.Provider for Gemini generateContent APIs.
type Provider struct {
	apiKey  string
	baseURL string
	client  *http.Client
	options map[string]string
	retries common.DefaultRetryPolicy
}

// New creates a new Gemini provider.
func New(apiKey, baseURL string, options map[string]string) (*Provider, error) {
	apiKey = common.ResolveAuthValue("GEMINI_API_KEY", apiKey, "")
	if apiKey == "" {
		return nil, fmt.Errorf("gemini: API key required (set GEMINI_API_KEY)")
	}
	baseURL = common.ResolveAuthValue("GEMINI_BASE_URL", baseURL, defaultBaseURL)

	return &Provider{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
		options: options,
		retries: common.NewDefaultRetryPolicy(),
	}, nil
}

func (p *Provider) Name() string            { return "gemini" }
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

// Chat sends a streaming request to the Gemini streamGenerateContent endpoint.
func (p *Provider) Chat(ctx context.Context, req types.ChatRequest) (<-chan types.StreamEvent, error) {
	body := p.buildRequest(req)

	for attempt := 0; ; attempt++ {
		httpReq, err := p.newRequest(ctx, req.Model, "streamGenerateContent", body)
		if err != nil {
			return nil, common.TranslateError("gemini", err)
		}

		resp, err := p.client.Do(httpReq)
		if err != nil {
			normalized := common.TranslateError("gemini", err)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("gemini", waitErr)
			}
			continue
		}

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			normalized := common.TranslateHTTPError("gemini", resp.StatusCode, respBody)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("gemini", waitErr)
			}
			continue
		}

		ch := make(chan types.StreamEvent, 64)
		go p.streamResponse(resp.Body, ch)
		return ch, nil
	}
}

// ChatSync sends a non-streaming generateContent request.
func (p *Provider) ChatSync(ctx context.Context, req types.ChatRequest) (*types.ChatResponse, error) {
	body := p.buildRequest(req)

	for attempt := 0; ; attempt++ {
		httpReq, err := p.newRequest(ctx, req.Model, "generateContent", body)
		if err != nil {
			return nil, common.TranslateError("gemini", err)
		}

		resp, err := p.client.Do(httpReq)
		if err != nil {
			normalized := common.TranslateError("gemini", err)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("gemini", waitErr)
			}
			continue
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, common.TranslateError("gemini", readErr)
		}

		if resp.StatusCode != http.StatusOK {
			normalized := common.TranslateHTTPError("gemini", resp.StatusCode, respBody)
			decision := p.retries.Decide(normalized, attempt)
			if !decision.Retry {
				return nil, normalized
			}
			if waitErr := common.WaitBackoff(ctx, decision.Backoff); waitErr != nil {
				return nil, common.TranslateError("gemini", waitErr)
			}
			continue
		}

		var apiResp geminiResponse
		if err := json.Unmarshal(respBody, &apiResp); err != nil {
			return nil, common.TranslateError("gemini", err)
		}

		return convertResponse(&apiResp), nil
	}
}

// ListModels returns available Gemini models via API listing endpoint.
func (p *Provider) ListModels(ctx context.Context) ([]types.Model, error) {
	endpoint := p.baseURL + "/" + apiVersion + "/models"
	q := url.Values{}
	q.Set("key", p.apiKey)
	endpoint = endpoint + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, common.TranslateError("gemini", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, common.TranslateError("gemini", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, common.TranslateError("gemini", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, common.TranslateHTTPError("gemini", resp.StatusCode, body)
	}

	var modelResp geminiModelsResponse
	if err := json.Unmarshal(body, &modelResp); err != nil {
		return nil, common.TranslateError("gemini", err)
	}

	models := make([]types.Model, 0, len(modelResp.Models))
	for _, m := range modelResp.Models {
		if !supportsGenerateContent(m.SupportedGenerationMethods) {
			continue
		}
		id := strings.TrimPrefix(m.Name, "models/")
		models = append(models, types.Model{
			ID:             id,
			Name:           id,
			Provider:       "gemini",
			ContextWindow:  m.InputTokenLimit,
			MaxOutput:      m.OutputTokenLimit,
			SupportsTools:  true,
			SupportsVision: true,
		})
	}

	return models, nil
}

type geminiGenerateRequest struct {
	Contents          []geminiContent    `json:"contents,omitempty"`
	Tools             []geminiToolHolder `json:"tools,omitempty"`
	SystemInstruction *geminiContent     `json:"systemInstruction,omitempty"`
	GenerationConfig  *generationConfig  `json:"generationConfig,omitempty"`
}

type generationConfig struct {
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

type geminiToolHolder struct {
	FunctionDeclarations []geminiFunctionDecl `json:"functionDeclarations,omitempty"`
}

type geminiFunctionDecl struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text             string                `json:"text,omitempty"`
	FunctionCall     *geminiFunctionCall   `json:"functionCall,omitempty"`
	FunctionResponse *geminiFunctionResult `json:"functionResponse,omitempty"`
}

type geminiFunctionCall struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args,omitempty"`
}

type geminiFunctionResult struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response,omitempty"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
	} `json:"usageMetadata"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason"`
}

type geminiModelsResponse struct {
	Models []struct {
		Name                       string   `json:"name"`
		InputTokenLimit            int      `json:"inputTokenLimit"`
		OutputTokenLimit           int      `json:"outputTokenLimit"`
		SupportedGenerationMethods []string `json:"supportedGenerationMethods"`
	} `json:"models"`
}

func (p *Provider) buildRequest(req types.ChatRequest) geminiGenerateRequest {
	apiReq := geminiGenerateRequest{}
	if req.System != "" {
		apiReq.SystemInstruction = &geminiContent{Parts: []geminiPart{{Text: req.System}}}
	}

	toolNameByUseID := make(map[string]string)
	for _, msg := range req.Messages {
		content := geminiContent{Role: mapRole(msg.Role)}
		for _, block := range msg.Content {
			switch block.Type {
			case types.ContentText:
				if block.Text != "" {
					content.Parts = append(content.Parts, geminiPart{Text: block.Text})
				}
			case types.ContentToolUse:
				toolNameByUseID[block.ToolUseID] = block.ToolName
				content.Parts = append(content.Parts, geminiPart{FunctionCall: &geminiFunctionCall{Name: block.ToolName, Args: block.Input}})
			case types.ContentToolResult:
				result := map[string]any{"content": block.Content, "is_error": block.IsError}
				content.Parts = append(content.Parts, geminiPart{FunctionResponse: &geminiFunctionResult{Name: toolNameByUseID[block.ForToolUseID], Response: result}})
			}
		}
		if len(content.Parts) > 0 {
			apiReq.Contents = append(apiReq.Contents, content)
		}
	}

	if len(req.Tools) > 0 {
		holder := geminiToolHolder{}
		for _, tool := range req.Tools {
			holder.FunctionDeclarations = append(holder.FunctionDeclarations, geminiFunctionDecl{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  schemaToMap(tool.InputSchema),
			})
		}
		apiReq.Tools = []geminiToolHolder{holder}
	}

	if req.MaxTokens > 0 || req.Temperature != nil || req.TopP != nil || len(req.StopSequences) > 0 {
		apiReq.GenerationConfig = &generationConfig{
			MaxOutputTokens: req.MaxTokens,
			Temperature:     req.Temperature,
			TopP:            req.TopP,
			StopSequences:   req.StopSequences,
		}
	}

	return apiReq
}

func (p *Provider) newRequest(ctx context.Context, model, action string, body geminiGenerateRequest) (*http.Request, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("gemini: marshaling request: %w", err)
	}

	resolvedModel := normalizeModel(model)
	endpoint := p.baseURL + "/" + apiVersion + "/models/" + resolvedModel + ":" + action
	q := url.Values{}
	q.Set("key", p.apiKey)
	if action == "streamGenerateContent" {
		q.Set("alt", "sse")
	}
	endpoint = endpoint + "?" + q.Encode()

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (p *Provider) streamResponse(body io.ReadCloser, ch chan<- types.StreamEvent) {
	defer close(ch)
	defer body.Close()

	ch <- types.StreamEvent{Type: types.StreamStart}

	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	toolCounter := 0
	stopReason := types.StopReason("")

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if strings.TrimSpace(data) == "" || data == "[DONE]" {
			continue
		}

		var chunk geminiResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			ch <- types.StreamEvent{Type: types.StreamError, Error: err}
			return
		}

		if len(chunk.Candidates) == 0 {
			continue
		}

		candidate := chunk.Candidates[0]
		if candidate.FinishReason != "" {
			stopReason = mapFinishReason(candidate.FinishReason)
		}

		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				ch <- types.StreamEvent{Type: types.StreamContentDelta, Delta: part.Text}
			}
			if part.FunctionCall != nil {
				toolCounter++
				toolID := fmt.Sprintf("gemini-tool-%d", toolCounter)
				ch <- types.StreamEvent{Type: types.StreamToolUseStart, ToolUseID: toolID, ToolName: part.FunctionCall.Name}
				if len(part.FunctionCall.Args) > 0 {
					ch <- types.StreamEvent{Type: types.StreamToolUseDelta, ToolUseID: toolID, ToolName: part.FunctionCall.Name, Delta: string(part.FunctionCall.Args)}
				}
				ch <- types.StreamEvent{Type: types.StreamToolUseDone, ToolUseID: toolID, ToolName: part.FunctionCall.Name}
			}
		}

		if chunk.UsageMetadata.PromptTokenCount > 0 || chunk.UsageMetadata.CandidatesTokenCount > 0 {
			ch <- types.StreamEvent{
				Usage: &types.Usage{
					InputTokens:  chunk.UsageMetadata.PromptTokenCount,
					OutputTokens: chunk.UsageMetadata.CandidatesTokenCount,
				},
				UsageCumulative: true,
				StopReason:      stopReason,
			}
		}
	}

	ch <- types.StreamEvent{Type: types.StreamMessageDone, StopReason: stopReason}
}

func convertResponse(resp *geminiResponse) *types.ChatResponse {
	result := &types.ChatResponse{Model: defaultModel, StopReason: types.StopUnknown}
	if len(resp.Candidates) == 0 {
		return result
	}

	candidate := resp.Candidates[0]
	msg := types.Message{Role: types.RoleAssistant}
	toolCounter := 0
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			msg.Content = append(msg.Content, types.ContentBlock{Type: types.ContentText, Text: part.Text})
		}
		if part.FunctionCall != nil {
			toolCounter++
			msg.Content = append(msg.Content, types.ContentBlock{
				Type:      types.ContentToolUse,
				ToolUseID: fmt.Sprintf("gemini-tool-%d", toolCounter),
				ToolName:  part.FunctionCall.Name,
				Input:     part.FunctionCall.Args,
			})
		}
	}

	result.Message = msg
	result.Usage = types.Usage{
		InputTokens:  resp.UsageMetadata.PromptTokenCount,
		OutputTokens: resp.UsageMetadata.CandidatesTokenCount,
	}
	result.StopReason = mapFinishReason(candidate.FinishReason)

	return result
}

func mapRole(role types.Role) string {
	switch role {
	case types.RoleAssistant:
		return "model"
	default:
		return "user"
	}
}

func mapFinishReason(reason string) types.StopReason {
	switch reason {
	case "STOP", "FINISH_REASON_UNSPECIFIED":
		return types.StopEndTurn
	case "MAX_TOKENS":
		return types.StopMaxTokens
	case "SAFETY", "BLOCKLIST", "PROHIBITED_CONTENT", "SPII":
		return types.StopContentFilter
	case "RECITATION":
		return types.StopRefusal
	default:
		return types.StopUnknown
	}
}

func normalizeModel(model string) string {
	if model == "" {
		return defaultModel
	}
	return strings.TrimPrefix(model, "models/")
}

func supportsGenerateContent(methods []string) bool {
	for _, method := range methods {
		if method == "generateContent" || method == "streamGenerateContent" {
			return true
		}
	}
	return false
}

func schemaToMap(schema types.ToolSchema) map[string]any {
	if schema.Type == "" {
		return nil
	}
	out := map[string]any{"type": schema.Type}
	if len(schema.Properties) > 0 {
		props := make(map[string]any, len(schema.Properties))
		for name, prop := range schema.Properties {
			props[name] = propertyToMap(prop)
		}
		out["properties"] = props
	}
	if len(schema.Required) > 0 {
		out["required"] = schema.Required
	}
	return out
}

func propertyToMap(prop types.PropertySchema) map[string]any {
	out := map[string]any{}
	if prop.Type != "" {
		out["type"] = prop.Type
	}
	if prop.Description != "" {
		out["description"] = prop.Description
	}
	if len(prop.Enum) > 0 {
		out["enum"] = prop.Enum
	}
	if prop.Default != nil {
		out["default"] = prop.Default
	}
	if prop.Items != nil {
		out["items"] = propertyToMap(*prop.Items)
	}
	if prop.Minimum != nil {
		out["minimum"] = *prop.Minimum
	}
	if prop.Maximum != nil {
		out["maximum"] = *prop.Maximum
	}
	return out
}
