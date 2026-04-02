package tools

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestWebFetchRuntimeMetadataFields(t *testing.T) {
	resp := &http.Response{StatusCode: 200, Status: "200 OK", Header: http.Header{"Content-Type": []string{"text/plain"}}}
	out := prependWebFetchMetadataWithRuntime("ok", "https://example.com", resp, "text", 123, 45*time.Millisecond)
	for _, needle := range []string{"final_url:", "bytes_read: 123", "duration_ms:"} {
		if !strings.Contains(out, needle) {
			t.Fatalf("missing %q in metadata: %s", needle, out)
		}
	}
}
