package tools

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/alliecatowo/alliecode/internal/mcp"
	"github.com/alliecatowo/alliecode/internal/types"
)

func TestMCPToolProxyExecute(t *testing.T) {
	proxy := &MCPToolProxy{
		serverName: "srv",
		def: types.ToolDef{
			Name:        "sum",
			Description: "adds",
			InputSchema: types.ToolSchema{Type: "object"},
		},
		client: fakeMCPCaller{},
	}

	in := json.RawMessage(`{"a":1}`)
	out, err := proxy.Execute(context.Background(), in, types.ToolContext{})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if out.IsError {
		t.Fatalf("expected success, got error: %s", out.Content)
	}
	if out.Content != "ok" {
		t.Fatalf("unexpected content: %q", out.Content)
	}
	if proxy.Name() != "mcp__srv__sum" {
		t.Fatalf("unexpected name: %s", proxy.Name())
	}
}

type fakeMCPCaller struct{}

func (fakeMCPCaller) CallTool(ctx context.Context, name string, input json.RawMessage) (string, error) {
	return "ok", nil
}

func TestRegistryMCPManagerServerStatusesNilManager(t *testing.T) {
	rm := &registryMCPManager{}
	if got := rm.ServerStatuses(); got != nil {
		t.Fatalf("expected nil statuses for nil manager, got %+v", got)
	}
}

func TestRegistryMCPManagerNilManager(t *testing.T) {
	rm := &registryMCPManager{}

	if _, err := rm.ListResources(context.Background()); err == nil {
		t.Fatal("expected error for nil ListResources")
	}
	if _, err := rm.ListResourcesForServer(context.Background(), "srv"); err == nil {
		t.Fatal("expected error for nil ListResourcesForServer")
	}
	if _, err := rm.ReadResourceCached(context.Background(), "srv", "mem://a"); err == nil {
		t.Fatal("expected error for nil ReadResourceCached")
	}
	if err := rm.AuthenticateLocal(context.Background(), "srv", "tok"); err == nil {
		t.Fatal("expected error for nil AuthenticateLocal")
	}
	if rm.Authenticated("srv") {
		t.Fatal("expected false for nil Authenticated")
	}
	if got := rm.AuthStatus("srv"); got != mcp.AuthStatusUnknown {
		t.Fatalf("expected unknown auth status, got %q", got)
	}
	if _, ok := rm.ServerStatus("srv"); ok {
		t.Fatal("expected missing status for nil manager")
	}
	if err := rm.SetAuthStateNeedsAuth("srv"); err == nil {
		t.Fatal("expected error for nil SetAuthStateNeedsAuth")
	}
	if err := rm.SetAuthStateAuthenticating("srv"); err == nil {
		t.Fatal("expected error for nil SetAuthStateAuthenticating")
	}
	if err := rm.SetAuthStateAuthenticated("srv"); err == nil {
		t.Fatal("expected error for nil SetAuthStateAuthenticated")
	}
	if err := rm.SetAuthStateFailed("srv"); err == nil {
		t.Fatal("expected error for nil SetAuthStateFailed")
	}
}
