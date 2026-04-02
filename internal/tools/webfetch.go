package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

// WebFetchTool fetches web content via HTTP GET.
type WebFetchTool struct{}

type webFetchInput struct {
	URL     string `json:"url"`
	Format  string `json:"format"`
	Timeout int    `json:"timeout"`
}

type crossHostRedirectError struct {
	OriginalURL string
	RedirectURL string
	StatusCode  int
}

func (e *crossHostRedirectError) Error() string {
	return buildCrossHostRedirectMessage(e.OriginalURL, e.RedirectURL, e.StatusCode)
}

func (t *WebFetchTool) Name() string { return "WebFetch" }

func (t *WebFetchTool) Description() string {
	return "Fetches content from a URL via HTTP GET. HTML tags are stripped for cleaner output."
}

func (t *WebFetchTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"url": {
				Type:        "string",
				Description: "The URL to fetch.",
			},
			"format": {
				Type:        "string",
				Description: "Output format: markdown (default), text, or html.",
				Enum:        []string{"markdown", "text", "html"},
			},
			"timeout": {
				Type:        "integer",
				Description: "Timeout in seconds. Defaults to 30, max 120.",
			},
		},
		Required: []string{"url"},
	}
}

// htmlTagRe matches HTML tags for stripping.
var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

// whitespaceRe matches multiple consecutive whitespace/newline chars.
var whitespaceRe = regexp.MustCompile(`\n{3,}`)

func (t *WebFetchTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	started := time.Now()
	var in webFetchInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}

	if in.URL == "" {
		return types.ToolResult{Content: "url is required", IsError: true}, nil
	}

	resolvedURL, err := normalizeWebFetchURL(in.URL)
	if err != nil {
		return types.ToolResult{Content: err.Error(), IsError: true}, nil
	}

	format := strings.ToLower(strings.TrimSpace(in.Format))
	if format == "" {
		format = "markdown"
	}
	if format != "markdown" && format != "text" && format != "html" {
		return types.ToolResult{Content: "format must be one of: markdown, text, html", IsError: true}, nil
	}

	timeout := resolveWebFetchTimeout(in.Timeout)

	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) == 0 {
				return nil
			}
			origin := via[0].URL
			status := 302
			if prevResp := via[len(via)-1].Response; prevResp != nil && prevResp.StatusCode > 0 {
				status = prevResp.StatusCode
			}
			if !strings.EqualFold(origin.Hostname(), req.URL.Hostname()) {
				return &crossHostRedirectError{
					OriginalURL: origin.String(),
					RedirectURL: req.URL.String(),
					StatusCode:  status,
				}
			}
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resolvedURL, nil)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error creating request: %v", err), IsError: true}, nil
	}
	req.Header.Set("User-Agent", "AllieCode/1.0")

	resp, err := client.Do(req)
	if err != nil {
		var redirectErr *crossHostRedirectError
		if errors.As(err, &redirectErr) {
			return types.ToolResult{Content: redirectErr.Error()}, nil
		}
		return types.ToolResult{Content: fmt.Sprintf("Error fetching URL: %v", err), IsError: true}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return types.ToolResult{
			Content: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
			IsError: true,
		}, nil
	}

	// Limit response body size to 1MB.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Error reading response: %v", err), IsError: true}, nil
	}

	content := string(body)
	content = formatWebFetchContent(content, resp.Header.Get("Content-Type"), format)
	content = prependWebFetchMetadataWithRuntime(content, resolvedURL, resp, format, len(body), time.Since(started))

	// Truncate if too long.
	const maxLen = 100000
	if len(content) > maxLen {
		content = content[:maxLen] + "\n... (truncated)"
	}

	return types.ToolResult{Content: content}, nil
}

func normalizeWebFetchURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("url is required")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid url: %v", err)
	}
	if u.Scheme == "" {
		return "", fmt.Errorf("url must start with http:// or https://")
	}
	if !strings.EqualFold(u.Scheme, "http") && !strings.EqualFold(u.Scheme, "https") {
		return "", fmt.Errorf("url must start with http:// or https://")
	}
	if strings.EqualFold(u.Scheme, "http") {
		u.Scheme = "https"
	}
	if u.Host == "" {
		return "", fmt.Errorf("invalid url: missing host")
	}
	return u.String(), nil
}

func formatWebFetchContent(content, contentType, format string) string {
	isHTML := strings.Contains(strings.ToLower(contentType), "text/html") || strings.Contains(strings.ToLower(content), "<html")
	if format == "html" {
		return content
	}
	if isHTML {
		return stripHTML(content)
	}
	return strings.TrimSpace(content)
}

func resolveWebFetchTimeout(seconds int) time.Duration {
	timeout := 30 * time.Second
	if seconds > 0 {
		timeout = time.Duration(seconds) * time.Second
	}
	if timeout > 120*time.Second {
		return 120 * time.Second
	}
	return timeout
}

func prependWebFetchMetadata(content, resolvedURL string, resp *http.Response, format string) string {
	return prependWebFetchMetadataWithRuntime(content, resolvedURL, resp, format, 0, 0)
}

func prependWebFetchMetadataWithRuntime(content, resolvedURL string, resp *http.Response, format string, bytesRead int, duration time.Duration) string {
	contentType := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if contentType == "" {
		contentType = "unknown"
	}
	finalURL := resolvedURL
	redirected := false
	if resp != nil && resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
		redirected = !strings.EqualFold(finalURL, resolvedURL)
	}
	meta := renderStructuredBlock("web_fetch_metadata", []structuredField{
		{Key: "url", Value: resolvedURL},
		{Key: "final_url", Value: finalURL},
		{Key: "redirected", Value: fmt.Sprintf("%t", redirected)},
		{Key: "status_code", Value: fmt.Sprintf("%d", resp.StatusCode)},
		{Key: "status_text", Value: resp.Status},
		{Key: "content_type", Value: contentType},
		{Key: "format", Value: format},
		{Key: "bytes_read", Value: fmt.Sprintf("%d", bytesRead)},
		{Key: "duration_ms", Value: fmt.Sprintf("%d", duration.Milliseconds())},
	})
	if strings.TrimSpace(content) == "" {
		return meta
	}
	return meta + "\n\n" + content
}

func buildCrossHostRedirectMessage(originalURL, redirectURL string, statusCode int) string {
	statusText := "Found"
	switch statusCode {
	case http.StatusMovedPermanently:
		statusText = "Moved Permanently"
	case http.StatusTemporaryRedirect:
		statusText = "Temporary Redirect"
	case http.StatusPermanentRedirect:
		statusText = "Permanent Redirect"
	}
	return fmt.Sprintf("REDIRECT DETECTED: The URL redirects to a different host.\n\nOriginal URL: %s\nRedirect URL: %s\nStatus: %d %s\n\nTo continue, run WebFetch again using the redirect URL.", originalURL, redirectURL, statusCode, statusText)
}

// stripHTML removes HTML tags and cleans up whitespace.
func stripHTML(s string) string {
	// Remove script and style blocks entirely.
	scriptRe := regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`)
	s = scriptRe.ReplaceAllString(s, "")
	styleRe := regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`)
	s = styleRe.ReplaceAllString(s, "")

	// Replace block-level tags with newlines.
	blockRe := regexp.MustCompile(`(?i)</(p|div|h[1-6]|li|tr|br|hr)[^>]*>`)
	s = blockRe.ReplaceAllString(s, "\n")

	// Strip remaining tags.
	s = htmlTagRe.ReplaceAllString(s, "")

	// Clean up excessive whitespace.
	s = whitespaceRe.ReplaceAllString(s, "\n\n")
	s = strings.TrimSpace(s)

	return s
}

func (t *WebFetchTool) IsReadOnly(input types.ToolInput) bool {
	return true
}

func (t *WebFetchTool) IsDestructive(input types.ToolInput) bool {
	return false
}

func (t *WebFetchTool) IsConcurrencySafe(input types.ToolInput) bool {
	return true
}

func (t *WebFetchTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyWeb, "webfetch", input, toolCtx, types.PermissionAsk)
}
