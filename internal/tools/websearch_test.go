package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestHTTPWebSearchProviderSearchAndLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("q"); got != "test query" {
			t.Fatalf("unexpected query: %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"title":"A","url":"https://a","snippet":"sa"},{"title":"B","url":"https://b","snippet":"sb"}]}`))
	}))
	defer server.Close()

	p := &HTTPWebSearchProvider{Endpoint: server.URL, ProviderName: "generic", Client: server.Client()}
	results, err := p.Search(context.Background(), "test query", 1)
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Title != "A" {
		t.Fatalf("unexpected result: %+v", results[0])
	}
}

func TestWebSearchToolUsesInjectedProvider(t *testing.T) {
	old := webSearchProvider
	t.Cleanup(func() { webSearchProvider = old })

	SetWebSearchProvider(&fakeSearchProvider{results: []webSearchResult{{Title: "Hello", URL: "https://example.com", Snippet: "snippet"}}})

	in, _ := json.Marshal(webSearchInput{Query: "x", NumResults: 5})
	res, err := (&WebSearchTool{}).Execute(context.Background(), in, types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected non-error result: %s", res.Content)
	}
	if !strings.Contains(res.Content, "Hello") {
		t.Fatalf("unexpected content: %s", res.Content)
	}
	if !strings.Contains(res.Content, "\"result_count\": 1") {
		t.Fatalf("expected contract json output, got: %s", res.Content)
	}
	if !strings.Contains(res.Content, "[web_search_metadata]") {
		t.Fatalf("expected metadata block, got: %s", res.Content)
	}
}

func TestWebSearchToolOutputContractWhenEmpty(t *testing.T) {
	old := webSearchProvider
	t.Cleanup(func() { webSearchProvider = old })

	SetWebSearchProvider(&fakeSearchProvider{})

	in, _ := json.Marshal(webSearchInput{Query: "x", NumResults: 5})
	res, err := (&WebSearchTool{}).Execute(context.Background(), in, types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if res.IsError {
		t.Fatalf("expected non-error result: %s", res.Content)
	}

	var out struct {
		Query       string `json:"query"`
		ResultCount int    `json:"result_count"`
		Results     []any  `json:"results"`
	}
	jsonContent := strings.SplitN(res.Content, "\n\n[web_search_metadata]", 2)[0]
	if err := json.Unmarshal([]byte(jsonContent), &out); err != nil {
		t.Fatalf("expected json output: %v\ncontent: %s", err, res.Content)
	}
	if out.Query != "x" || out.ResultCount != 0 || len(out.Results) != 0 {
		t.Fatalf("unexpected contract output: %+v", out)
	}
}

func TestCountWebSearchResults(t *testing.T) {
	if got := countWebSearchResults(`{"result_count":3}`); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
	if got := countWebSearchResults("nope"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}

	meta := renderStructuredBlock("web_search_metadata", []structuredField{{Key: "returned_results", Value: strconv.Itoa(3)}})
	if !strings.Contains(meta, "returned_results: 3") {
		t.Fatalf("unexpected metadata block: %s", meta)
	}
}

func TestWebSearchToolClampsNumResults(t *testing.T) {
	old := webSearchProvider
	t.Cleanup(func() { webSearchProvider = old })

	spy := &fakeSearchProvider{results: []webSearchResult{{Title: "A", URL: "https://a"}}}
	SetWebSearchProvider(spy)

	in, _ := json.Marshal(webSearchInput{Query: "x", NumResults: 1000})
	_, err := (&WebSearchTool{}).Execute(context.Background(), in, types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if spy.lastNumResults != 50 {
		t.Fatalf("expected num_results clamp to 50, got %d", spy.lastNumResults)
	}
}

type fakeSearchProvider struct {
	results        []webSearchResult
	err            error
	lastNumResults int
}

func (f *fakeSearchProvider) Search(ctx context.Context, query string, numResults int) ([]webSearchResult, error) {
	f.lastNumResults = numResults
	return f.results, f.err
}
