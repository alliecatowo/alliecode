package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/alliecatowo/alliecode/internal/mcp"
	"github.com/alliecatowo/alliecode/internal/types"
)

// MCPManager defines the MCP integration surface used by tools registry.
type MCPManager interface {
	Tools() []types.Tool
	Call(ctx context.Context, serverName, toolName string, input json.RawMessage) (string, error)
	ListResources(ctx context.Context) ([]mcp.Resource, error)
	ListResourcesForServer(ctx context.Context, serverName string) ([]mcp.Resource, error)
	ReadResourceCached(ctx context.Context, serverName, uri string) (mcp.ResourceContent, error)
	ServerStatuses() []mcp.ServerStatus
	AuthenticateLocal(ctx context.Context, serverName, token string) error
	Authenticated(serverName string) bool
	AuthStatus(serverName string) mcp.AuthStatus
	ServerStatus(serverName string) (mcp.ServerStatus, bool)
	SetAuthStateNeedsAuth(serverName string) error
	SetAuthStateAuthenticating(serverName string) error
	SetAuthStateAuthenticated(serverName string) error
	SetAuthStateFailed(serverName string) error
	Close() error
}

type registryMCPManager struct {
	mgr   *mcp.Manager
	tools []types.Tool
}

// NewMCPManager initializes MCP clients and exposes MCP tools.
func NewMCPManager(ctx context.Context, configs []mcp.ServerConfig) (MCPManager, error) {
	mgr, err := mcp.NewManager(ctx, configs)
	if err != nil {
		return nil, err
	}

	wrapped := &registryMCPManager{mgr: mgr}
	for _, remote := range mgr.Tools() {
		wrapped.tools = append(wrapped.tools, &MCPToolProxy{
			serverName: remote.ServerName,
			def:        remote.Def,
			client:     mcpToolCaller{mgr: mgr, serverName: remote.ServerName},
		})
	}

	return wrapped, nil
}

func (m *registryMCPManager) Tools() []types.Tool {
	out := make([]types.Tool, len(m.tools))
	copy(out, m.tools)
	return out
}

func (m *registryMCPManager) Close() error {
	if m.mgr == nil {
		return nil
	}
	return m.mgr.Close()
}

func (m *registryMCPManager) Call(ctx context.Context, serverName, toolName string, input json.RawMessage) (string, error) {
	if m.mgr == nil {
		return "", fmt.Errorf("mcp manager is not configured")
	}
	return m.mgr.Call(ctx, serverName, toolName, input)
}

func (m *registryMCPManager) ListResources(ctx context.Context) ([]mcp.Resource, error) {
	if m.mgr == nil {
		return nil, fmt.Errorf("mcp manager is not configured")
	}
	return m.mgr.ListResources(ctx)
}

func (m *registryMCPManager) ReadResourceCached(ctx context.Context, serverName, uri string) (mcp.ResourceContent, error) {
	if m.mgr == nil {
		return mcp.ResourceContent{}, fmt.Errorf("mcp manager is not configured")
	}
	return m.mgr.ReadResourceCached(ctx, serverName, uri)
}

func (m *registryMCPManager) ListResourcesForServer(ctx context.Context, serverName string) ([]mcp.Resource, error) {
	if m.mgr == nil {
		return nil, fmt.Errorf("mcp manager is not configured")
	}
	return m.mgr.ListResourcesForServer(ctx, serverName)
}

func (m *registryMCPManager) AuthenticateLocal(ctx context.Context, serverName, token string) error {
	if m.mgr == nil {
		return fmt.Errorf("mcp manager is not configured")
	}
	return m.mgr.AuthenticateLocal(ctx, serverName, token)
}

func (m *registryMCPManager) Authenticated(serverName string) bool {
	if m.mgr == nil {
		return false
	}
	return m.mgr.Authenticated(serverName)
}

func (m *registryMCPManager) AuthStatus(serverName string) mcp.AuthStatus {
	if m.mgr == nil {
		return mcp.AuthStatusUnknown
	}
	return m.mgr.AuthStatus(serverName)
}

func (m *registryMCPManager) ServerStatus(serverName string) (mcp.ServerStatus, bool) {
	if m.mgr == nil {
		return mcp.ServerStatus{}, false
	}
	return m.mgr.ServerStatus(serverName)
}

func (m *registryMCPManager) ServerStatuses() []mcp.ServerStatus {
	if m.mgr == nil {
		return nil
	}
	return m.mgr.ServerStatuses()
}

func (m *registryMCPManager) SetAuthStateNeedsAuth(serverName string) error {
	if m.mgr == nil {
		return fmt.Errorf("mcp manager is not configured")
	}
	return m.mgr.SetAuthStateNeedsAuth(serverName)
}

func (m *registryMCPManager) SetAuthStateAuthenticating(serverName string) error {
	if m.mgr == nil {
		return fmt.Errorf("mcp manager is not configured")
	}
	return m.mgr.SetAuthStateAuthenticating(serverName)
}

func (m *registryMCPManager) SetAuthStateAuthenticated(serverName string) error {
	if m.mgr == nil {
		return fmt.Errorf("mcp manager is not configured")
	}
	return m.mgr.SetAuthStateAuthenticated(serverName)
}

func (m *registryMCPManager) SetAuthStateFailed(serverName string) error {
	if m.mgr == nil {
		return fmt.Errorf("mcp manager is not configured")
	}
	return m.mgr.SetAuthStateFailed(serverName)
}

// MCPToolProxy adapts a remote MCP tool as a local tool interface.
type MCPToolProxy struct {
	serverName string
	def        types.ToolDef
	client     mcpCaller
}

type mcpCaller interface {
	CallTool(ctx context.Context, name string, input json.RawMessage) (string, error)
}

type mcpToolCaller struct {
	mgr        *mcp.Manager
	serverName string
}

func (c mcpToolCaller) CallTool(ctx context.Context, name string, input json.RawMessage) (string, error) {
	if c.mgr == nil {
		return "", fmt.Errorf("mcp manager is not configured")
	}
	_ = c.mgr.SetAuthStateAuthenticating(c.serverName)
	out, err := c.mgr.Call(ctx, c.serverName, name, input)
	if err != nil {
		if mcp.IsNeedsAuthError(err) {
			_ = c.mgr.SetAuthStateNeedsAuth(c.serverName)
		} else {
			_ = c.mgr.SetAuthStateFailed(c.serverName)
		}
		return "", err
	}
	_ = c.mgr.SetAuthStateAuthenticated(c.serverName)
	return out, nil
}

func (t *MCPToolProxy) Name() string {
	return fmt.Sprintf("mcp__%s__%s", t.serverName, t.def.Name)
}

func (t *MCPToolProxy) Description() string {
	return t.def.Description
}

func (t *MCPToolProxy) InputSchema() types.ToolSchema {
	return t.def.InputSchema
}

func (t *MCPToolProxy) Execute(ctx context.Context, input types.ToolInput, toolCtx types.ToolContext) (types.ToolResult, error) {
	content, err := t.client.CallTool(ctx, t.def.Name, json.RawMessage(input))
	if err != nil {
		return types.ToolResult{Content: err.Error(), IsError: true}, nil
	}
	return types.ToolResult{Content: content}, nil
}

func (t *MCPToolProxy) IsReadOnly(input types.ToolInput) bool {
	return false
}

func (t *MCPToolProxy) IsDestructive(input types.ToolInput) bool {
	return false
}

func (t *MCPToolProxy) IsConcurrencySafe(input types.ToolInput) bool {
	return true
}

func (t *MCPToolProxy) CheckPermissions(input types.ToolInput, toolCtx types.ToolContext) types.ToolPermission {
	return evaluateToolPermissionByFamily(toolFamilyMCP, t.Name(), input, toolCtx, types.PermissionAsk)
}
