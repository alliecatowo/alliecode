package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/types"
)

func TestWebSearchMetadataIncludesProviderAndDuration(t *testing.T) {
	old := webSearchProvider
	t.Cleanup(func() { webSearchProvider = old })
	SetWebSearchProvider(&fakeSearchProvider{results: []webSearchResult{{Title: "A", URL: "https://a"}}})
	in, _ := json.Marshal(webSearchInput{Query: "x"})
	res, err := (&WebSearchTool{}).Execute(context.Background(), in, types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("search failed: err=%v content=%q", err, res.Content)
	}
	if !strings.Contains(res.Content, "provider:") || !strings.Contains(res.Content, "duration_ms:") {
		t.Fatalf("expected provider/duration metadata: %s", res.Content)
	}
}
