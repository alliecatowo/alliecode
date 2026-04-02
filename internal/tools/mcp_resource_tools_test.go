package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/alliecatowo/alliecode/internal/mcp"
	"github.com/alliecatowo/alliecode/internal/types"
)

type fakeRegistryMCPManager struct {
	resources      []mcp.Resource
	content        mcp.ResourceContent
	toolResult     string
	authenticated  map[string]bool
	authStatus     map[string]mcp.AuthStatus
	connection     map[string]mcp.ServerConnectionState
	transport      map[string]mcp.TransportType
	listErr        error
	listServerErr  error
	readErr        error
	callErr        error
	authErr        error
	lastReadServer string
	lastReadURI    string
	lastListServer string
	lastCallServer string
	lastCallTool   string
	lastCallInput  json.RawMessage
}

func (f *fakeRegistryMCPManager) Tools() []types.Tool { return nil }
func (f *fakeRegistryMCPManager) Close() error        { return nil }
func (f *fakeRegistryMCPManager) Call(ctx context.Context, serverName, toolName string, input json.RawMessage) (string, error) {
	f.lastCallServer = serverName
	f.lastCallTool = toolName
	f.lastCallInput = input
	if f.callErr != nil {
		return "", f.callErr
	}
	if f.toolResult != "" {
		return f.toolResult, nil
	}
	return "ok", nil
}
func (f *fakeRegistryMCPManager) ListResources(ctx context.Context) ([]mcp.Resource, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := make([]mcp.Resource, len(f.resources))
	copy(out, f.resources)
	return out, nil
}
func (f *fakeRegistryMCPManager) ListResourcesForServer(ctx context.Context, serverName string) ([]mcp.Resource, error) {
	f.lastListServer = serverName
	if f.listServerErr != nil {
		return nil, f.listServerErr
	}
	if tt, ok := f.transport[serverName]; ok && tt != mcp.TransportStdio {
		return nil, fmt.Errorf("%w: %s", mcp.ErrTransportUnsupported, tt)
	}
	out := make([]mcp.Resource, 0, len(f.resources))
	for _, r := range f.resources {
		if r.ServerName == serverName {
			out = append(out, r)
		}
	}
	return out, nil
}
func (f *fakeRegistryMCPManager) ReadResourceCached(ctx context.Context, serverName, uri string) (mcp.ResourceContent, error) {
	f.lastReadServer = serverName
	f.lastReadURI = uri
	if f.readErr != nil {
		return mcp.ResourceContent{}, f.readErr
	}
	if tt, ok := f.transport[serverName]; ok && tt != mcp.TransportStdio {
		return mcp.ResourceContent{}, fmt.Errorf("%w: %s", mcp.ErrTransportUnsupported, tt)
	}
	return f.content, nil
}
func (f *fakeRegistryMCPManager) AuthenticateLocal(ctx context.Context, serverName, token string) error {
	if f.authErr != nil {
		return f.authErr
	}
	if f.authenticated == nil {
		f.authenticated = map[string]bool{}
	}
	f.authenticated[serverName] = true
	return nil
}
func (f *fakeRegistryMCPManager) Authenticated(serverName string) bool {
	return f.authenticated[serverName]
}
func (f *fakeRegistryMCPManager) AuthStatus(serverName string) mcp.AuthStatus {
	if f.authStatus == nil {
		return mcp.AuthStatusUnknown
	}
	return f.authStatus[serverName]
}
func (f *fakeRegistryMCPManager) ServerStatus(serverName string) (mcp.ServerStatus, bool) {
	tt, ok := f.transport[serverName]
	if !ok {
		return mcp.ServerStatus{}, false
	}
	return mcp.ServerStatus{
		ServerName:      serverName,
		TransportType:   tt,
		ConnectionState: f.connection[serverName],
		Authenticated:   f.authenticated[serverName],
		AuthStatus:      f.authStatus[serverName],
	}, true
}
func (f *fakeRegistryMCPManager) ServerStatuses() []mcp.ServerStatus {
	statuses := make([]mcp.ServerStatus, 0, len(f.transport))
	for serverName, tt := range f.transport {
		statuses = append(statuses, mcp.ServerStatus{
			ServerName:      serverName,
			TransportType:   tt,
			ConnectionState: f.connection[serverName],
			Authenticated:   f.authenticated[serverName],
			AuthStatus:      f.authStatus[serverName],
		})
	}
	return statuses
}
func (f *fakeRegistryMCPManager) SetAuthStateNeedsAuth(serverName string) error {
	if f.authenticated == nil {
		f.authenticated = map[string]bool{}
	}
	if f.authStatus == nil {
		f.authStatus = map[string]mcp.AuthStatus{}
	}
	if f.connection == nil {
		f.connection = map[string]mcp.ServerConnectionState{}
	}
	f.authenticated[serverName] = false
	f.authStatus[serverName] = mcp.AuthStatusUnauthenticated
	f.connection[serverName] = mcp.ServerConnectionNeedsAuth
	return nil
}
func (f *fakeRegistryMCPManager) SetAuthStateAuthenticating(serverName string) error {
	if f.authenticated == nil {
		f.authenticated = map[string]bool{}
	}
	if f.authStatus == nil {
		f.authStatus = map[string]mcp.AuthStatus{}
	}
	if f.connection == nil {
		f.connection = map[string]mcp.ServerConnectionState{}
	}
	f.authenticated[serverName] = false
	f.authStatus[serverName] = mcp.AuthStatusAuthenticating
	f.connection[serverName] = mcp.ServerConnectionPending
	return nil
}
func (f *fakeRegistryMCPManager) SetAuthStateAuthenticated(serverName string) error {
	if f.authenticated == nil {
		f.authenticated = map[string]bool{}
	}
	if f.authStatus == nil {
		f.authStatus = map[string]mcp.AuthStatus{}
	}
	if f.connection == nil {
		f.connection = map[string]mcp.ServerConnectionState{}
	}
	f.authenticated[serverName] = true
	f.authStatus[serverName] = mcp.AuthStatusAuthenticated
	f.connection[serverName] = mcp.ServerConnectionConnected
	return nil
}
func (f *fakeRegistryMCPManager) SetAuthStateFailed(serverName string) error {
	if f.authenticated == nil {
		f.authenticated = map[string]bool{}
	}
	if f.authStatus == nil {
		f.authStatus = map[string]mcp.AuthStatus{}
	}
	if f.connection == nil {
		f.connection = map[string]mcp.ServerConnectionState{}
	}
	f.authenticated[serverName] = false
	f.authStatus[serverName] = mcp.AuthStatusFailed
	f.connection[serverName] = mcp.ServerConnectionFailed
	return nil
}

func TestMCPResourceListToolDeterministicOutput(t *testing.T) {
	fake := &fakeRegistryMCPManager{resources: []mcp.Resource{
		{ServerName: "b", URI: "mem://2", Name: "two"},
		{ServerName: "a", URI: "mem://3", Name: "three"},
		{ServerName: "a", URI: "mem://1", Name: "one"},
	}, transport: map[string]mcp.TransportType{"a": mcp.TransportStdio, "b": mcp.TransportStdio}, authStatus: map[string]mcp.AuthStatus{"a": mcp.AuthStatusAuthenticated, "b": mcp.AuthStatusUnauthenticated}, connection: map[string]mcp.ServerConnectionState{"a": mcp.ServerConnectionConnected, "b": mcp.ServerConnectionNeedsAuth}, authenticated: map[string]bool{"a": true}}

	res, err := (&MCPResourceListTool{manager: fake}).Execute(context.Background(), []byte(`{}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("resource list failed: err=%v content=%q", err, res.Content)
	}

	var listed struct {
		Resources      []mcp.Resource     `json:"resources"`
		Total          int                `json:"total"`
		ServerStatuses []mcp.ServerStatus `json:"server_statuses"`
	}
	if err := json.Unmarshal([]byte(res.Content), &listed); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if listed.Total != 3 || len(listed.Resources) != 3 {
		t.Fatalf("unexpected totals: %+v", listed)
	}
	if listed.Resources[0].ServerName != "a" || listed.Resources[0].URI != "mem://1" || listed.Resources[2].ServerName != "b" {
		t.Fatalf("unexpected resource ordering: %+v", listed.Resources)
	}
	if len(listed.ServerStatuses) != 2 || listed.ServerStatuses[0].ServerName != "a" || listed.ServerStatuses[1].ServerName != "b" {
		t.Fatalf("unexpected server statuses: %+v", listed.ServerStatuses)
	}

	filtered, err := (&MCPResourceListTool{manager: fake}).Execute(context.Background(), []byte(`{"server_name":"a"}`), types.ToolContext{})
	if err != nil || filtered.IsError {
		t.Fatalf("filtered resource list failed: err=%v content=%q", err, filtered.Content)
	}
	if err := json.Unmarshal([]byte(filtered.Content), &listed); err != nil {
		t.Fatalf("invalid filtered JSON output: %v", err)
	}
	if listed.Total != 2 || len(listed.Resources) != 2 {
		t.Fatalf("unexpected filtered totals: %+v", listed)
	}
}

func TestMCPResourceReadAndAuthTools(t *testing.T) {
	fake := &fakeRegistryMCPManager{
		content:       mcp.ResourceContent{URI: "mem://x", MIMEType: "text/plain", Text: "hello"},
		authenticated: map[string]bool{"srv": true},
		authStatus:    map[string]mcp.AuthStatus{"srv": mcp.AuthStatusAuthenticated},
		connection:    map[string]mcp.ServerConnectionState{"srv": mcp.ServerConnectionConnected},
		transport:     map[string]mcp.TransportType{"srv": mcp.TransportStdio},
	}

	readRes, err := (&MCPResourceReadTool{manager: fake}).Execute(context.Background(), []byte(`{"server_name":"srv","uri":"mem://x"}`), types.ToolContext{})
	if err != nil || readRes.IsError {
		t.Fatalf("resource read failed: err=%v content=%q", err, readRes.Content)
	}
	if fake.lastReadServer != "srv" || fake.lastReadURI != "mem://x" {
		t.Fatalf("unexpected read call args: server=%q uri=%q", fake.lastReadServer, fake.lastReadURI)
	}
	const expectedRead = `{"server_name":"srv","authenticated":true,"auth_status":"authenticated","connection_state":"connected","content":{"URI":"mem://x","MIMEType":"text/plain","Text":"hello"}}`
	if strings.Contains(readRes.Content, expectedRead) {
		t.Fatalf("read output unexpectedly used legacy shape: %s", readRes.Content)
	}
	var readOut map[string]any
	if err := json.Unmarshal([]byte(readRes.Content), &readOut); err != nil {
		t.Fatalf("unexpected read output JSON error: %v", err)
	}
	if readOut["server_name"] != "srv" || readOut["transport"] != string(mcp.TransportStdio) {
		t.Fatalf("unexpected read output: %+v", readOut)
	}

	authRes, err := (&MCPAuthLocalTool{manager: fake}).Execute(context.Background(), []byte(`{"server_name":"srv","token":"tok"}`), types.ToolContext{})
	if err != nil || authRes.IsError {
		t.Fatalf("auth tool failed: err=%v content=%q", err, authRes.Content)
	}
	const expectedAuth = `{"server_name":"srv","authenticated":true}`
	if authRes.Content != expectedAuth {
		t.Fatalf("unexpected auth output:\nwant: %s\ngot:  %s", expectedAuth, authRes.Content)
	}
}

func TestMCPResourceToolsErrorPaths(t *testing.T) {
	nilMgrRead, err := (&MCPResourceReadTool{}).Execute(context.Background(), []byte(`{"server_name":"srv","uri":"mem://x"}`), types.ToolContext{})
	if err != nil || !nilMgrRead.IsError {
		t.Fatalf("expected nil manager error for read: err=%v content=%q", err, nilMgrRead.Content)
	}

	fake := &fakeRegistryMCPManager{listErr: errors.New("boom"), readErr: errors.New("bad read"), authErr: errors.New("denied"), transport: map[string]mcp.TransportType{"srv": mcp.TransportStdio}}

	listRes, err := (&MCPResourceListTool{manager: fake}).Execute(context.Background(), []byte(`{}`), types.ToolContext{})
	if err != nil || !listRes.IsError {
		t.Fatalf("expected list error: err=%v content=%q", err, listRes.Content)
	}

	badInput, err := (&MCPAuthLocalTool{manager: fake}).Execute(context.Background(), []byte(`{"server_name":"srv"}`), types.ToolContext{})
	if err != nil || !badInput.IsError {
		t.Fatalf("expected missing token validation error: err=%v content=%q", err, badInput.Content)
	}

	readRes, err := (&MCPResourceReadTool{manager: fake}).Execute(context.Background(), []byte(`{"server_name":"srv","uri":"mem://x"}`), types.ToolContext{})
	if err != nil || !readRes.IsError {
		t.Fatalf("expected read error: err=%v content=%q", err, readRes.Content)
	}

	authRes, err := (&MCPAuthLocalTool{manager: fake}).Execute(context.Background(), []byte(`{"server_name":"srv","token":"tok"}`), types.ToolContext{})
	if err != nil || !authRes.IsError {
		t.Fatalf("expected auth error: err=%v content=%q", err, authRes.Content)
	}

	var out map[string]any
	if err := json.Unmarshal([]byte(authRes.Content), &out); err != nil {
		t.Fatalf("expected deterministic JSON error, got %q (%v)", authRes.Content, err)
	}
	if out["code"] != "MCP_AUTH_FAILED" {
		t.Fatalf("unexpected auth error payload: %+v", out)
	}
}

func TestMCPResourceToolsDeterministicUnsupportedTransportFailures(t *testing.T) {
	fake := &fakeRegistryMCPManager{
		resources: []mcp.Resource{{ServerName: "sse", URI: "mem://x", Name: "x"}},
		transport: map[string]mcp.TransportType{"sse": mcp.TransportSSE},
	}

	listRes, err := (&MCPResourceListTool{manager: fake}).Execute(context.Background(), []byte(`{"server_name":"sse"}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("resource list error: %v", err)
	}
	if !listRes.IsError {
		t.Fatalf("unexpected list unsupported success: %+v", listRes)
	}
	var listErr map[string]any
	if err := json.Unmarshal([]byte(listRes.Content), &listErr); err != nil {
		t.Fatalf("expected JSON list error, got %q", listRes.Content)
	}
	if listErr["code"] != "MCP_TRANSPORT_UNSUPPORTED" {
		t.Fatalf("unexpected list unsupported result: %+v", listErr)
	}

	readRes, err := (&MCPResourceReadTool{manager: fake}).Execute(context.Background(), []byte(`{"server_name":"sse","uri":"mem://x"}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("resource read error: %v", err)
	}
	if !readRes.IsError {
		t.Fatalf("unexpected read unsupported success: %+v", readRes)
	}
	var readErr map[string]any
	if err := json.Unmarshal([]byte(readRes.Content), &readErr); err != nil {
		t.Fatalf("expected JSON read error, got %q", readRes.Content)
	}
	if readErr["code"] != "MCP_TRANSPORT_UNSUPPORTED" {
		t.Fatalf("unexpected read unsupported result: %+v", readErr)
	}
}

func TestMCPResourceReadRequiresAuth(t *testing.T) {
	fake := &fakeRegistryMCPManager{
		transport:  map[string]mcp.TransportType{"srv": mcp.TransportStdio},
		authStatus: map[string]mcp.AuthStatus{"srv": mcp.AuthStatusUnauthenticated},
		connection: map[string]mcp.ServerConnectionState{"srv": mcp.ServerConnectionNeedsAuth},
	}

	res, err := (&MCPResourceReadTool{manager: fake}).Execute(context.Background(), []byte(`{"server_name":"srv","uri":"mem://x"}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("resource read error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("unexpected auth gate success: %+v", res)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(res.Content), &out); err != nil {
		t.Fatalf("expected JSON auth gate error, got %q", res.Content)
	}
	if out["code"] != "MCP_AUTH_REQUIRED" {
		t.Fatalf("unexpected auth gate result: %+v", out)
	}
}

func TestMCPToolInvokeToolSuccessAndNeedsAuth(t *testing.T) {
	fake := &fakeRegistryMCPManager{
		transport:     map[string]mcp.TransportType{"srv": mcp.TransportStdio},
		authStatus:    map[string]mcp.AuthStatus{"srv": mcp.AuthStatusAuthenticated},
		connection:    map[string]mcp.ServerConnectionState{"srv": mcp.ServerConnectionConnected},
		authenticated: map[string]bool{"srv": true},
		toolResult:    "done",
	}
	tool := &MCPToolInvokeTool{manager: fake}

	res, err := tool.Execute(context.Background(), []byte(`{"server_name":"srv","tool_name":"sum","arguments":{"a":1}}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("invoke success failed: err=%v content=%q", err, res.Content)
	}
	if fake.lastCallServer != "srv" || fake.lastCallTool != "sum" {
		t.Fatalf("unexpected call routing: server=%q tool=%q", fake.lastCallServer, fake.lastCallTool)
	}
	var success map[string]any
	if err := json.Unmarshal([]byte(res.Content), &success); err != nil {
		t.Fatalf("expected invoke success JSON, got %q", res.Content)
	}
	if success["server_name"] != "srv" || success["transport"] != string(mcp.TransportStdio) {
		t.Fatalf("unexpected invoke output: %+v", success)
	}

	fake.callErr = errors.New("401 unauthorized")
	res, err = tool.Execute(context.Background(), []byte(`{"server_name":"srv","tool_name":"sum"}`), types.ToolContext{})
	if err != nil {
		t.Fatalf("invoke auth error execution failed: %v", err)
	}
	if !res.IsError {
		t.Fatalf("unexpected needs-auth invoke success: %+v", res)
	}
	var needsAuth map[string]any
	if err := json.Unmarshal([]byte(res.Content), &needsAuth); err != nil {
		t.Fatalf("expected needs-auth JSON error, got %q", res.Content)
	}
	if needsAuth["code"] != "MCP_AUTH_REQUIRED" {
		t.Fatalf("unexpected needs-auth invoke output: %+v", needsAuth)
	}
}

func TestMCPAuthStatusTool(t *testing.T) {
	fake := &fakeRegistryMCPManager{
		authenticated: map[string]bool{"srv": true},
		authStatus:    map[string]mcp.AuthStatus{"srv": mcp.AuthStatusAuthenticated},
		connection:    map[string]mcp.ServerConnectionState{"srv": mcp.ServerConnectionConnected},
		transport:     map[string]mcp.TransportType{"srv": mcp.TransportStdio},
	}

	res, err := (&MCPAuthStatusTool{manager: fake}).Execute(context.Background(), []byte(`{"server_name":"srv"}`), types.ToolContext{})
	if err != nil || res.IsError {
		t.Fatalf("auth status tool failed: err=%v content=%q", err, res.Content)
	}
	if !strings.Contains(res.Content, `"server_name":"srv"`) || !strings.Contains(res.Content, `"status":`) {
		t.Fatalf("unexpected auth status output: %s", res.Content)
	}

	missing, err := (&MCPAuthStatusTool{manager: fake}).Execute(context.Background(), []byte(`{"server_name":"missing"}`), types.ToolContext{})
	if err != nil || !missing.IsError {
		t.Fatalf("expected missing server error: err=%v content=%q", err, missing.Content)
	}
}
