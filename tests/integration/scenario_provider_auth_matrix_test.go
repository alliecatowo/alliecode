package integration_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/providers/anthropic"
	"github.com/alliecatowo/alliecode/internal/providers/common"
	"github.com/alliecatowo/alliecode/internal/providers/gemini"
	"github.com/alliecatowo/alliecode/internal/providers/ollama"
	"github.com/alliecatowo/alliecode/internal/providers/openai"
	"github.com/alliecatowo/alliecode/internal/types"
)

type providerAuthCase struct {
	Name              string `json:"name"`
	Kind              string `json:"kind"`
	Provider          string `json:"provider"`
	StatusCode        int    `json:"status_code"`
	Body              string `json:"body"`
	WantClass         string `json:"want_class"`
	WantRetryable     bool   `json:"want_retryable"`
	WantErrorContains string `json:"want_error_contains"`
}

func TestScenarioMatrix_ProviderAuthAndRateLimitMapping(t *testing.T) {
	paths := loadScenarioMatrixPaths(t, "provider_auth", "*.json")
	for _, path := range paths {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			var tc providerAuthCase
			decodeScenarioCase(t, path, &tc)

			switch tc.Kind {
			case "missing_key":
				err := runProviderMissingKeyCase(t, tc.Provider)
				if err == nil {
					t.Fatalf("expected constructor error for provider %q", tc.Provider)
				}
				if tc.WantErrorContains != "" && !strings.Contains(err.Error(), tc.WantErrorContains) {
					t.Fatalf("error = %q, want contains %q", err.Error(), tc.WantErrorContains)
				}
			case "http_error_mapping":
				err := common.TranslateHTTPError(tc.Provider, tc.StatusCode, []byte(tc.Body))
				assertNormalizedProviderError(t, tc, err)
			case "chat_sync_stub_error":
				err := runProviderChatSyncStubErrorCase(t, tc)
				if err == nil {
					t.Fatalf("expected ChatSync() error for provider %q", tc.Provider)
				}
				assertNormalizedProviderError(t, tc, err)
			default:
				t.Fatalf("unknown scenario kind %q", tc.Kind)
			}
		})
	}
}

func runProviderChatSyncStubErrorCase(t *testing.T, tc providerAuthCase) error {
	t.Helper()

	statusCode := tc.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusUnauthorized
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(tc.Body))
	}))
	t.Cleanup(server.Close)

	req := types.ChatRequest{
		Model:     "does-not-exist",
		MaxTokens: 8,
		Messages: []types.Message{
			types.NewTextMessage(types.RoleUser, "ping"),
		},
	}

	ctx := context.Background()
	switch tc.Provider {
	case "openai":
		p, err := openai.New("fixture-key", server.URL, "", nil)
		if err != nil {
			t.Fatalf("openai.New() error = %v", err)
		}
		_, err = p.ChatSync(ctx, req)
		return err
	case "gemini":
		p, err := gemini.New("fixture-key", server.URL, nil)
		if err != nil {
			t.Fatalf("gemini.New() error = %v", err)
		}
		_, err = p.ChatSync(ctx, req)
		return err
	case "anthropic":
		p, err := anthropic.New("fixture-key", "", "", server.URL, nil)
		if err != nil {
			t.Fatalf("anthropic.New() error = %v", err)
		}
		_, err = p.ChatSync(ctx, req)
		return err
	case "ollama":
		p, err := ollama.New(server.URL, nil)
		if err != nil {
			t.Fatalf("ollama.New() error = %v", err)
		}
		_, err = p.ChatSync(ctx, req)
		return err
	default:
		t.Fatalf("chat_sync_stub_error scenario does not support provider %q", tc.Provider)
		return nil
	}
}

func assertNormalizedProviderError(t *testing.T, tc providerAuthCase, err error) {
	t.Helper()

	nerr, ok := err.(*common.NormalizedError)
	if !ok {
		t.Fatalf("error type = %T, want *common.NormalizedError", err)
	}
	if string(nerr.Class) != tc.WantClass {
		t.Fatalf("normalized class = %q, want %q", nerr.Class, tc.WantClass)
	}
	if nerr.Retryable != tc.WantRetryable {
		t.Fatalf("normalized retryable = %t, want %t", nerr.Retryable, tc.WantRetryable)
	}
	if tc.WantErrorContains != "" && !strings.Contains(strings.ToLower(nerr.Error()), strings.ToLower(tc.WantErrorContains)) {
		t.Fatalf("normalized error = %q, want contains %q", nerr.Error(), tc.WantErrorContains)
	}
}

func runProviderMissingKeyCase(t *testing.T, provider string) error {
	t.Helper()

	switch provider {
	case "openai":
		t.Setenv("OPENAI_API_KEY", "")
		_, err := openai.New("", "", "", nil)
		return err
	case "gemini":
		t.Setenv("GEMINI_API_KEY", "")
		_, err := gemini.New("", "", nil)
		return err
	case "anthropic":
		t.Setenv("ANTHROPIC_API_KEY", "")
		t.Setenv("ANTHROPIC_AUTH_TOKEN", "")
		t.Setenv("ANTHROPIC_ACCESS_TOKEN", "")
		_, err := anthropic.New("", "", "", "", nil)
		return err
	default:
		t.Fatalf("missing_key scenario does not support provider %q", provider)
		return nil
	}
}
