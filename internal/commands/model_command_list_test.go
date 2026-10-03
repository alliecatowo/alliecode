package commands

import (
	"context"
	"strings"
	"testing"
)

func TestModelListByProvider(t *testing.T) {
	cmd := NewModelCommand()
	state := &RuntimeState{ProviderName: "openai"}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"list", "openai"}})
	if err != nil || !strings.Contains(res.Message, "MODEL_LIST") {
		t.Fatalf("model list failed: %v %q", err, res.Message)
	}
	if !strings.Contains(res.Message, "provider=openai") || !strings.Contains(res.Message, "provider_ready=false") {
		t.Fatalf("expected provider/readiness in list output: %q", res.Message)
	}
}

func TestModelListAllMatrixIncludesProvidersAndEntries(t *testing.T) {
	cmd := NewModelCommand()
	state := &RuntimeState{}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"list", "all"}})
	if err != nil {
		t.Fatalf("model list all failed: %v", err)
	}
	if !strings.Contains(res.Message, "MODEL_LIST_ALL") {
		t.Fatalf("expected MODEL_LIST_ALL output, got %q", res.Message)
	}
	for _, providerLine := range []string{"provider.1=anthropic", "provider.2=gemini", "provider.3=ollama", "provider.4=openai"} {
		if !strings.Contains(res.Message, providerLine) {
			t.Fatalf("expected provider matrix line %q in %q", providerLine, res.Message)
		}
	}
}

func TestModelListWithoutProviderShowsGroupedProviders(t *testing.T) {
	cmd := NewModelCommand()
	state := &RuntimeState{LoggedIn: true}
	res, err := cmd.Execute(context.Background(), Context{State: state}, Invocation{Name: "model", Args: []string{"list"}})
	if err != nil {
		t.Fatalf("model list grouped failed: %v", err)
	}
	if !strings.Contains(res.Message, "MODEL_LIST") || !strings.Contains(res.Message, "scope=providers") {
		t.Fatalf("expected grouped provider listing, got %q", res.Message)
	}
}
