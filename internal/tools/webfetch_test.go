package tools

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNormalizeWebFetchURLUpgradesHTTP(t *testing.T) {
	out, err := normalizeWebFetchURL("http://example.com/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "https://example.com/path" {
		t.Fatalf("expected https upgrade, got %q", out)
	}
}

func TestNormalizeWebFetchURLRejectsInvalid(t *testing.T) {
	if _, err := normalizeWebFetchURL("ftp://example.com"); err == nil {
		t.Fatalf("expected invalid scheme error")
	}
	if _, err := normalizeWebFetchURL("https://"); err == nil {
		t.Fatalf("expected missing host error")
	}
}

func TestResolveWebFetchTimeoutBounds(t *testing.T) {
	if got := resolveWebFetchTimeout(0); got != 30*time.Second {
		t.Fatalf("expected default timeout, got %v", got)
	}
	if got := resolveWebFetchTimeout(999); got != 120*time.Second {
		t.Fatalf("expected capped timeout, got %v", got)
	}
}

func TestFormatWebFetchContentDefaults(t *testing.T) {
	html := "<html><body><h1>Hello</h1></body></html>"
	out := formatWebFetchContent(html, "text/html", "markdown")
	if strings.Contains(out, "<h1>") {
		t.Fatalf("expected stripped html in markdown mode, got %q", out)
	}
	out = formatWebFetchContent(html, "text/html", "html")
	if out != html {
		t.Fatalf("expected raw html in html mode")
	}
}

func TestWebFetchPrependsMetadataBlock(t *testing.T) {
	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
	}
	out := prependWebFetchMetadata("hello", "https://example.com", resp, "text")
	if !strings.Contains(out, "[web_fetch_metadata]") {
		t.Fatalf("expected metadata block, got: %s", out)
	}
	if !strings.Contains(out, "status_code: 200") {
		t.Fatalf("expected status code metadata, got: %s", out)
	}
}

func TestWebFetchReturnsRedirectContractForCrossHost(t *testing.T) {
	msg := buildCrossHostRedirectMessage("https://a.example", "https://b.example", http.StatusFound)
	if !strings.Contains(msg, "REDIRECT DETECTED") {
		t.Fatalf("expected redirect contract, got: %s", msg)
	}
}

func TestWebFetchContractFormattingHelpers(t *testing.T) {
	html := "<html><body><h1>Title</h1><p>Hello</p></body></html>"
	formatted := formatWebFetchContent(html, "text/html", "markdown")
	if strings.Contains(formatted, "<h1>") {
		t.Fatalf("expected stripped HTML, got: %s", formatted)
	}
	resp := &http.Response{StatusCode: 200, Status: "200 OK", Header: http.Header{"Content-Type": []string{"text/html"}}}
	out := prependWebFetchMetadata(formatted, "https://example.com", resp, "markdown")
	if !strings.Contains(out, "[web_fetch_metadata]") {
		t.Fatalf("expected metadata block, got: %s", out)
	}
}
