package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/alliecatowo/alliecode/internal/types"
)

// WebSearchTool performs web searches using a configurable provider.
type WebSearchTool struct{}

type webSearchInput struct {
	Query      string `json:"query"`
	NumResults int    `json:"num_results"`
}

type webSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

type WebSearchProvider interface {
	Search(ctx context.Context, query string, numResults int) ([]webSearchResult, error)
}

type HTTPWebSearchProvider struct {
	ProviderName string
	Endpoint     string
	APIKey       string
	Client       *http.Client
}

var webSearchProvider WebSearchProvider

// SetWebSearchProvider overrides the provider used by WebSearchTool.
func SetWebSearchProvider(p WebSearchProvider) {
	webSearchProvider = p
}

func (t *WebSearchTool) Name() string { return "WebSearch" }

func (t *WebSearchTool) Description() string {
	return "Performs a web search and returns ranked results from a configured provider."
}

func (t *WebSearchTool) InputSchema() types.ToolSchema {
	return types.ToolSchema{
		Type: "object",
		Properties: map[string]types.PropertySchema{
			"query": {
				Type:        "string",
				Description: "The search query.",
			},
			"num_results": {
				Type:        "integer",
				Description: "Number of results to return. Default is 10.",
			},
		},
		Required: []string{"query"},
	}
}

func (t *WebSearchTool) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	_ = toolCtx
	started := time.Now()
	var in webSearchInput
	if err := json.Unmarshal(input, &in); err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Invalid input: %v", err), IsError: true}, nil
	}

	if in.Query == "" {
		return types.ToolResult{Content: "query is required", IsError: true}, nil
	}
	if in.NumResults <= 0 {
		in.NumResults = 10
	}
	if in.NumResults > 50 {
		in.NumResults = 50
	}

	provider, err := resolveWebSearchProvider()
	if err != nil {
		return types.ToolResult{Content: err.Error(), IsError: true}, nil
	}

	results, err := provider.Search(ctx, in.Query, in.NumResults)
	if err != nil {
		return types.ToolResult{Content: fmt.Sprintf("Web search failed: %v", err), IsError: true}, nil
	}

	providerName := "custom"
	if httpProvider, ok := provider.(*HTTPWebSearchProvider); ok {
		providerName = strings.TrimSpace(httpProvider.ProviderName)
		if providerName == "" {
			providerName = "generic"
		}
	}

	jsonOut := formatWebSearchResults(in.Query, results)
	meta := renderStructuredBlock("web_search_metadata", []structuredField{
		{Key: "query", Value: strings.TrimSpace(in.Query)},
		{Key: "requested_results", Value: fmt.Sprintf("%d", in.NumResults)},
		{Key: "returned_results", Value: fmt.Sprintf("%d", countWebSearchResults(jsonOut))},
		{Key: "provider", Value: providerName},
		{Key: "duration_ms", Value: fmt.Sprintf("%d", time.Since(started).Milliseconds())},
	})
	return types.ToolResult{Content: jsonOut + "\n\n" + meta}, nil

}

func countWebSearchResults(content string) int {
	var out struct {
		ResultCount int `json:"result_count"`
	}
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return 0
	}
	return out.ResultCount
}

func resolveWebSearchProvider() (WebSearchProvider, error) {
	if webSearchProvider != nil {
		return webSearchProvider, nil
	}

	endpoint := strings.TrimSpace(os.Getenv("ALLIECODE_WEBSEARCH_ENDPOINT"))
	if endpoint == "" {
		return nil, fmt.Errorf("web search provider not configured: set ALLIECODE_WEBSEARCH_ENDPOINT or install a provider")
	}

	provider := strings.TrimSpace(os.Getenv("ALLIECODE_WEBSEARCH_PROVIDER"))
	if provider == "" {
		provider = "generic"
	}

	apiKey := strings.TrimSpace(os.Getenv("ALLIECODE_WEBSEARCH_API_KEY"))

	return &HTTPWebSearchProvider{
		ProviderName: provider,
		Endpoint:     endpoint,
		APIKey:       apiKey,
		Client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}, nil
}

func (p *HTTPWebSearchProvider) Search(ctx context.Context, query string, numResults int) ([]webSearchResult, error) {
	if p.Endpoint == "" {
		return nil, fmt.Errorf("search endpoint is empty")
	}

	u, err := url.Parse(p.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint: %w", err)
	}
	q := u.Query()
	q.Set("q", query)
	q.Set("num_results", fmt.Sprintf("%d", numResults))
	q.Set("provider", p.ProviderName)
	u.RawQuery = q.Encode()

	client := p.Client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Results []webSearchResult `json:"results"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2*1024*1024)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("invalid search response: %w", err)
	}

	results := payload.Results
	if len(results) > numResults {
		results = results[:numResults]
	}
	return results, nil
}

func formatWebSearchResults(query string, results []webSearchResult) string {
	type webSearchItem struct {
		Rank    int    `json:"rank"`
		Title   string `json:"title"`
		URL     string `json:"url"`
		Snippet string `json:"snippet,omitempty"`
	}
	type webSearchOutput struct {
		Query       string          `json:"query"`
		ResultCount int             `json:"result_count"`
		Results     []webSearchItem `json:"results"`
	}

	sanitized := make([]webSearchResult, 0, len(results))
	for _, r := range results {
		r.Title = strings.TrimSpace(r.Title)
		r.URL = strings.TrimSpace(r.URL)
		r.Snippet = strings.TrimSpace(r.Snippet)
		if r.Title == "" && r.URL == "" && r.Snippet == "" {
			continue
		}
		sanitized = append(sanitized, r)
	}

	items := make([]webSearchItem, 0, len(sanitized))
	for i, r := range sanitized {
		title := r.Title
		if title == "" {
			title = "(untitled)"
		}
		items = append(items, webSearchItem{
			Rank:    i + 1,
			Title:   title,
			URL:     r.URL,
			Snippet: r.Snippet,
		})
	}

	out := webSearchOutput{
		Query:       strings.TrimSpace(query),
		ResultCount: len(items),
		Results:     items,
	}
	encoded, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Sprintf("{\"query\":%q,\"result_count\":0,\"results\":[]}", strings.TrimSpace(query))
	}
	return string(encoded)
}

func (t *WebSearchTool) IsReadOnly(input types.ToolInput) bool {
	return true
}

func (t *WebSearchTool) IsDestructive(input types.ToolInput) bool {
	return false
}

func (t *WebSearchTool) IsConcurrencySafe(input types.ToolInput) bool {
	return true
}

func (t *WebSearchTool) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyWeb, "websearch", input, toolCtx, types.PermissionAllowed)
}
