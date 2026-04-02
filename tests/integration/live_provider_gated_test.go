package integration_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alliecatowo/alliecode/internal/providers/anthropic"
	"github.com/alliecatowo/alliecode/internal/providers/common"
	"github.com/alliecatowo/alliecode/internal/providers/gemini"
	"github.com/alliecatowo/alliecode/internal/providers/ollama"
	"github.com/alliecatowo/alliecode/internal/providers/openai"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestLiveProvider_OpenAI_ChatSyncAndStream_Gated(t *testing.T) {
	requireLiveProviderLane(t)
	requireNonEmptyEnv(t, "OPENAI_API_KEY")

	p, err := openai.New("", "", "", nil)
	if err != nil {
		t.Fatalf("openai.New() error = %v", err)
	}

	req := minimalLiveChatRequest(openAIIntegrationModel())

	ctxSync, cancelSync := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancelSync()
	resp, err := p.ChatSync(ctxSync, req)
	if err != nil {
		t.Fatalf("ChatSync() error = %v", err)
	}
	assertLiveSyncShape(t, resp)

	ctxStream, cancelStream := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancelStream()
	stream, err := p.Chat(ctxStream, req)
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	assertLiveStreamShape(t, stream)
}

func TestLiveProvider_OpenAI_ToolPromptFallback_Gated(t *testing.T) {
	requireLiveProviderLane(t)
	requireNonEmptyEnv(t, "OPENAI_API_KEY")

	p, err := openai.New("", "", "", nil)
	if err != nil {
		t.Fatalf("openai.New() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	resp, err := p.ChatSync(ctx, toolCapableLiveChatRequest(openAIIntegrationModel()))
	if err != nil {
		t.Fatalf("ChatSync() error = %v", err)
	}
	assertToolCapableFallbackShape(t, resp)
}

func TestLiveProvider_Anthropic_ChatSyncAndStream_Gated(t *testing.T) {
	requireLiveProviderLane(t)
	requireAnyNonEmptyEnv(t, "ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_ACCESS_TOKEN")

	p, err := anthropic.New("", "", "", "", nil)
	if err != nil {
		t.Fatalf("anthropic.New() error = %v", err)
	}

	req := minimalLiveChatRequest(anthropicIntegrationModel())

	ctxSync, cancelSync := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancelSync()
	resp, err := p.ChatSync(ctxSync, req)
	if err != nil {
		t.Fatalf("ChatSync() error = %v", err)
	}
	assertLiveSyncShape(t, resp)

	ctxStream, cancelStream := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancelStream()
	stream, err := p.Chat(ctxStream, req)
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	assertLiveStreamShape(t, stream)
}

func TestLiveProvider_Anthropic_ToolPromptFallback_Gated(t *testing.T) {
	requireLiveProviderLane(t)
	requireAnyNonEmptyEnv(t, "ANTHROPIC_API_KEY", "ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_ACCESS_TOKEN")

	p, err := anthropic.New("", "", "", "", nil)
	if err != nil {
		t.Fatalf("anthropic.New() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	resp, err := p.ChatSync(ctx, toolCapableLiveChatRequest(anthropicIntegrationModel()))
	if err != nil {
		t.Fatalf("ChatSync() error = %v", err)
	}
	assertToolCapableFallbackShape(t, resp)
}

func TestLiveProvider_Gemini_ChatSyncAndStream_Gated(t *testing.T) {
	requireLiveProviderLane(t)
	requireNonEmptyEnv(t, "GEMINI_API_KEY")

	p, err := gemini.New("", "", nil)
	if err != nil {
		t.Fatalf("gemini.New() error = %v", err)
	}

	req := minimalLiveChatRequest(geminiIntegrationModel())

	ctxSync, cancelSync := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancelSync()
	resp, err := p.ChatSync(ctxSync, req)
	if err != nil {
		t.Fatalf("ChatSync() error = %v", err)
	}
	assertLiveSyncShape(t, resp)

	ctxStream, cancelStream := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancelStream()
	stream, err := p.Chat(ctxStream, req)
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	assertLiveStreamShape(t, stream)
}

func TestLiveProvider_Gemini_ToolPromptFallback_Gated(t *testing.T) {
	requireLiveProviderLane(t)
	requireNonEmptyEnv(t, "GEMINI_API_KEY")

	p, err := gemini.New("", "", nil)
	if err != nil {
		t.Fatalf("gemini.New() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	resp, err := p.ChatSync(ctx, toolCapableLiveChatRequest(geminiIntegrationModel()))
	if err != nil {
		t.Fatalf("ChatSync() error = %v", err)
	}
	assertToolCapableFallbackShape(t, resp)
}

func TestLiveProvider_Ollama_ChatSyncAndStream_Gated(t *testing.T) {
	requireLiveProviderLane(t)
	requireNonEmptyEnv(t, "AC_IT_OLLAMA")

	p, err := ollama.New("", nil)
	if err != nil {
		t.Fatalf("ollama.New() error = %v", err)
	}

	req := minimalLiveChatRequest(ollamaLiveIntegrationModel())

	ctxSync, cancelSync := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancelSync()
	resp, err := p.ChatSync(ctxSync, req)
	if err != nil {
		t.Fatalf("ChatSync() error = %v", err)
	}
	assertLiveSyncShape(t, resp)

	ctxStream, cancelStream := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancelStream()
	stream, err := p.Chat(ctxStream, req)
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	assertLiveStreamShape(t, stream)
}

func TestLiveProvider_Ollama_ToolPromptFallback_Gated(t *testing.T) {
	requireLiveProviderLane(t)
	requireNonEmptyEnv(t, "AC_IT_OLLAMA")

	p, err := ollama.New("", nil)
	if err != nil {
		t.Fatalf("ollama.New() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	resp, err := p.ChatSync(ctx, toolCapableLiveChatRequest(ollamaLiveIntegrationModel()))
	if err != nil {
		t.Fatalf("ChatSync() error = %v", err)
	}
	assertToolCapableFallbackShape(t, resp)
}

func TestLiveProvider_Classification_AuthAndRateLimit_Gated(t *testing.T) {
	requireLiveProviderLane(t)

	type tc struct {
		name          string
		provider      string
		statusCode    int
		body          string
		wantClass     common.ErrorClass
		wantRetryable bool
	}

	cases := []tc{
		{name: "openai_auth", provider: "openai", statusCode: 401, body: `{"error":{"message":"invalid api key"}}`, wantClass: common.ErrorClassAuth, wantRetryable: false},
		{name: "openai_rate_limit", provider: "openai", statusCode: 429, body: `{"error":{"message":"too many requests"}}`, wantClass: common.ErrorClassRateLimit, wantRetryable: true},
		{name: "anthropic_auth", provider: "anthropic", statusCode: 401, body: `{"type":"error","error":{"type":"authentication_error"}}`, wantClass: common.ErrorClassAuth, wantRetryable: false},
		{name: "anthropic_rate_limit", provider: "anthropic", statusCode: 429, body: `{"type":"error","error":{"type":"rate_limit_error"}}`, wantClass: common.ErrorClassRateLimit, wantRetryable: true},
		{name: "gemini_auth", provider: "gemini", statusCode: 401, body: `{"error":{"message":"invalid key"}}`, wantClass: common.ErrorClassAuth, wantRetryable: false},
		{name: "gemini_rate_limit", provider: "gemini", statusCode: 429, body: `{"error":{"message":"quota exceeded"}}`, wantClass: common.ErrorClassRateLimit, wantRetryable: true},
		{name: "ollama_auth", provider: "ollama", statusCode: 401, body: `{"error":"unauthorized"}`, wantClass: common.ErrorClassAuth, wantRetryable: false},
		{name: "ollama_rate_limit", provider: "ollama", statusCode: 429, body: `{"error":"too many requests"}`, wantClass: common.ErrorClassRateLimit, wantRetryable: true},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := common.TranslateHTTPError(tc.provider, tc.statusCode, []byte(tc.body))
			nerr, ok := err.(*common.NormalizedError)
			if !ok {
				t.Fatalf("error type = %T, want *common.NormalizedError", err)
			}
			if nerr.Class != tc.wantClass {
				t.Fatalf("class = %q, want %q", nerr.Class, tc.wantClass)
			}
			if nerr.Retryable != tc.wantRetryable {
				t.Fatalf("retryable = %t, want %t", nerr.Retryable, tc.wantRetryable)
			}
		})
	}
}

func minimalLiveChatRequest(model string) types.ChatRequest {
	temp := 0.0
	return types.ChatRequest{
		Model:       model,
		Temperature: &temp,
		MaxTokens:   24,
		Messages: []types.Message{
			types.NewTextMessage(types.RoleUser, "Reply with exactly: integration-ok"),
		},
	}
}

func toolCapableLiveChatRequest(model string) types.ChatRequest {
	temp := 0.0
	return types.ChatRequest{
		Model:       model,
		Temperature: &temp,
		MaxTokens:   48,
		Messages: []types.Message{
			types.NewTextMessage(types.RoleUser, "Use add(a,b) for 2 and 3 when possible, then answer with the total."),
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
	}
}

func assertLiveSyncShape(t *testing.T, resp *types.ChatResponse) {
	t.Helper()
	if resp == nil {
		t.Fatalf("chat response is nil")
	}
	if resp.Message.Role != types.RoleAssistant {
		t.Fatalf("response role = %q, want %q", resp.Message.Role, types.RoleAssistant)
	}
	if strings.TrimSpace(resp.Message.GetText()) == "" && len(resp.Message.GetToolUses()) == 0 {
		t.Fatalf("response text and tool calls are both empty")
	}
}

func assertToolCapableFallbackShape(t *testing.T, resp *types.ChatResponse) {
	t.Helper()
	assertLiveSyncShape(t, resp)
	toolUses := resp.Message.GetToolUses()
	if len(toolUses) == 0 {
		if strings.TrimSpace(resp.Message.GetText()) == "" {
			t.Fatalf("tool-capable fallback response text is empty")
		}
		return
	}
	for i, tu := range toolUses {
		if strings.TrimSpace(tu.ToolUseID) == "" {
			t.Fatalf("toolUses[%d].ToolUseID is empty", i)
		}
		if strings.TrimSpace(tu.ToolName) == "" {
			t.Fatalf("toolUses[%d].ToolName is empty", i)
		}
	}
}

func assertLiveStreamShape(t *testing.T, stream <-chan types.StreamEvent) {
	t.Helper()

	var sawStart, sawDone bool
	for ev := range stream {
		if ev.Type == types.StreamStart {
			sawStart = true
		}
		if ev.Type == types.StreamMessageDone {
			sawDone = true
			if ev.Message == nil {
				t.Fatalf("message_done event has nil message")
			}
			if strings.TrimSpace(ev.Message.GetText()) == "" && len(ev.Message.GetToolUses()) == 0 {
				t.Fatalf("message_done has empty text and no tool calls")
			}
		}
	}

	if !sawStart {
		t.Fatalf("stream did not emit %q", types.StreamStart)
	}
	if !sawDone {
		t.Fatalf("stream did not emit %q", types.StreamMessageDone)
	}
}

func requireLiveProviderLane(t *testing.T) {
	t.Helper()
	requireNonEmptyEnv(t, "AC_IT_LIVE_PROVIDERS")
}

func requireNonEmptyEnv(t *testing.T, key string) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		t.Skipf("skipping live provider test: %s is not set", key)
	}
	return v
}

func requireAnyNonEmptyEnv(t *testing.T, keys ...string) {
	t.Helper()
	for _, key := range keys {
		if strings.TrimSpace(os.Getenv(key)) != "" {
			return
		}
	}
	t.Skipf("skipping live provider test: set one of %s", strings.Join(keys, ", "))
}

func openAIIntegrationModel() string {
	if model := strings.TrimSpace(os.Getenv("AC_IT_OPENAI_MODEL")); model != "" {
		return model
	}
	return "gpt-4o-mini"
}

func anthropicIntegrationModel() string {
	if model := strings.TrimSpace(os.Getenv("AC_IT_ANTHROPIC_MODEL")); model != "" {
		return model
	}
	return "claude-haiku-3-5-20241022"
}

func geminiIntegrationModel() string {
	if model := strings.TrimSpace(os.Getenv("AC_IT_GEMINI_MODEL")); model != "" {
		return model
	}
	return "gemini-2.5-flash"
}

func ollamaLiveIntegrationModel() string {
	if model := strings.TrimSpace(os.Getenv("AC_IT_OLLAMA_MODEL")); model != "" {
		return model
	}
	return "llama3"
}
